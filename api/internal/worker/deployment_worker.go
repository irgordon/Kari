package worker

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"kari/api/internal/core/domain"
	"kari/api/internal/core/services"
	"kari/api/proto/agent" // Generated gRPC client
)

type DeploymentWorker struct {
	repo       domain.DeploymentRepository
	crypto     services.CryptoService
	agent      agent.SystemAgentClient
	pollInterval time.Duration
}

func NewDeploymentWorker(
	repo domain.DeploymentRepository,
	crypto services.CryptoService,
	agent agent.SystemAgentClient,
) *DeploymentWorker {
	return &DeploymentWorker{
		repo:         repo,
		crypto:       crypto,
		agent:        agent,
		pollInterval: 5 * time.Second,
	}
}

// Start initiates the background polling loop
func (w *DeploymentWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processNextTask(ctx)
		}
	}
}

func (w *DeploymentWorker) processNextTask(ctx context.Context) {
	// 1. 🛡️ Claim a pending deployment (Optimistic Locking)
	deployment, err := w.repo.ClaimNextPending(ctx)
	if err != nil || deployment == nil {
		return // No tasks or error
	}

	log.Printf("🏗️ Kari Panel: Starting deployment for %s", deployment.DomainName)

	// 2. 🛡️ Zero-Trust: Decrypt the SSH Key for the transient gRPC call
	// The key is decrypted here and passed to the Agent; it is never stored in the database in plaintext.
	var sshKey string
	if deployment.EncryptedSSHKey != "" {
		// AssociatedData binds the secret to this specific AppID
		decrypted, err := w.crypto.Decrypt(ctx, deployment.EncryptedSSHKey, []byte(deployment.AppID))
		if err != nil {
			w.failDeployment(ctx, deployment, fmt.Errorf("crypto: failed to decrypt deploy key: %w", err))
			return
		}
		sshKey = string(decrypted)
	}

	// 3. 📡 Initiate gRPC Stream with the Muscle
	stream, err := w.agent.StreamDeployment(ctx, &agent.DeployRequest{
		AppId:        deployment.AppID,
		DomainName:   deployment.DomainName,
		RepoUrl:      deployment.RepoURL,
		Branch:       deployment.Branch,
		BuildCommand: deployment.BuildCommand,
		Port:         int32(deployment.TargetPort),
		SshKey:       &sshKey, // Transferred securely
		TraceId:      deployment.ID,
	})
	if err != nil {
		w.failDeployment(ctx, deployment, fmt.Errorf("rpc: agent connection failure: %w", err))
		return
	}

	// 4. 🚰 Pipe Logs & Monitor Progress
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			w.failDeployment(ctx, deployment, fmt.Errorf("rpc: stream broken: %w", err))
			return
		}

		// Append logs to the database or broadcast to the Kari UI
		_ = w.repo.AppendLog(ctx, deployment.ID, chunk.Content)
	}

	// 5. ✅ Finalize Success
	_ = w.repo.UpdateStatus(ctx, deployment.ID, domain.StatusSuccess)
	log.Printf("✅ Kari Panel: Deployment successful for %s", deployment.DomainName)
}

func (w *DeploymentWorker) failDeployment(ctx context.Context, d *domain.Deployment, err error) {
	log.Printf("❌ Kari Panel: Deployment failed: %v", err)
	_ = w.repo.AppendLog(ctx, d.ID, fmt.Sprintf("\n❌ ERROR: %v\n", err))
	_ = w.repo.UpdateStatus(ctx, d.ID, domain.StatusFailed)
}
