// Modules
mod app;
mod models;
mod services;
mod ui;

fn main() -> iced::Result {
    iced::run("EnvMan", app::update, app::view)
}
