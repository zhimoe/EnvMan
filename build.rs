#[cfg(target_os = "windows")]
fn main() {
    let mut res = winres::WindowsResource::new();
    res.set_manifest_file("resources/manifest.xml");
    res.compile().expect("Failed to compile Windows resources");
}

#[cfg(not(target_os = "windows"))]
fn main() {
    println!("only supports windows build")
}
