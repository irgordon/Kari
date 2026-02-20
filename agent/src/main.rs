use kari_agent::app;

fn main() {
    if let Err(err) = app::run() {
        eprintln!("kari-agent failed: {err}");
        std::process::exit(1);
    }
}
