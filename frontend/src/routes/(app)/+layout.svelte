<script lang="ts">
    import { page } from '$app/stores';
    import type { LayoutData } from './$types';

    // This data is strictly typed and securely provided by +layout.server.ts
    export let data: LayoutData;
    $: user = data.user;

    // Standardized navigation map
    const navItems = [
        { name: 'Dashboard', path: '/dashboard', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6' },
        { name: 'Applications', path: '/applications', icon: 'M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4' },
        { name: 'Domains', path: '/domains', icon: 'M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9' },
        { name: 'Action Center', path: '/action-center', icon: 'M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9' },
        { name: 'Settings', path: '/settings', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z' }
    ];

    // Helper to determine active state
    $: isActive = (path: string) => $page.url.pathname.startsWith(path);
</script>

<div class="flex h-screen w-full bg-kari-light-gray font-body antialiased text-kari-text">
    
    <aside class="w-64 flex flex-col bg-white border-r border-kari-warm-gray/20 shadow-sm flex-shrink-0">
        <div class="h-16 flex items-center px-6 border-b border-kari-warm-gray/10">
            <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded bg-kari-teal flex items-center justify-center text-white font-sans font-bold text-lg shadow-sm">
                    K
                </div>
                <span class="font-sans font-semibold tracking-tight text-xl text-kari-text">Karı</span>
            </div>
        </div>

        <nav class="flex-1 px-4 py-6 space-y-1 overflow-y-auto">
            {#each navItems as item}
                <a 
                    href={item.path}
                    class="flex items-center gap-3 px-3 py-2.5 rounded-md transition-all duration-200 font-medium
                           {isActive(item.path) 
                               ? 'bg-kari-teal/10 text-kari-teal' 
                               : 'text-kari-text hover:bg-kari-light-gray hover:text-kari-teal'}"
                >
                    <svg 
                        xmlns="http://www.w3.org/2000/svg" 
                        class="h-5 w-5 {isActive(item.path) ? 'text-kari-teal' : 'text-kari-warm-gray'}" 
                        fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
                    >
                        <path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
                    </svg>
                    {item.name}
                </a>
            {each}
        </nav>

        {#if user}
            <div class="p-4 border-t border-kari-warm-gray/20 bg-gray-50/50">
                <div class="flex items-center gap-3">
                    <div class="w-9 h-9 rounded-full bg-kari-teal/20 flex items-center justify-center text-kari-teal font-sans font-bold">
                        {user.email.charAt(0).toUpperCase()}
                    </div>
                    <div class="flex flex-col overflow-hidden">
                        <span class="text-sm font-semibold text-kari-text truncate">{user.email}</span>
                        <span class="text-xs text-kari-warm-gray font-medium">{user.roleName}</span>
                    </div>
                </div>
                
                <form method="POST" action="/logout" class="mt-3">
                    <button 
                        type="submit" 
                        class="w-full text-left text-xs font-medium text-kari-warm-gray hover:text-kari-teal transition-colors flex items-center gap-2 px-1"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                        </svg>
                        Sign out
                    </button>
                </form>
            </div>
        {/if}
    </aside>

    <main class="flex-1 flex flex-col min-w-0 overflow-hidden">
        <header class="h-16 flex items-center px-8 border-b border-kari-warm-gray/10 bg-white shadow-sm shrink-0">
            <h1 class="font-sans font-semibold text-kari-text text-lg capitalize">
                {$page.url.pathname.split('/')[1] || 'Dashboard'}
            </h1>
        </header>

        <div class="flex-1 overflow-y-auto p-8">
            <div class="max-w-6xl mx-auto">
                <slot />
            </div>
        </div>
    </main>

</div>
