use iced::widget::{column, row, container};
use iced::{Length, Element, Task};

use crate::models::state::AppState;
use crate::models::message::Message;

use crate::ui::{tool_bar, tab_bar, variable_list_panel, editor_panel};

/// Application state
#[derive(Default)]
pub struct EnvManApp {
    pub state: AppState,
}

/// Initialize the application
pub fn init() -> EnvManApp {
    EnvManApp::default()
}

/// Update the application state
pub fn update(app: &mut EnvManApp, message: Message) -> Task<Message> {
    // Handle special messages before state update
    match &message {
        Message::RestartAsAdmin => {
            #[cfg(target_os = "windows")]
            {
                if let Err(e) = crate::services::restart_as_admin() {
                    return Task::perform(async { e.to_string() }, Message::Error);
                }
                return Task::none();
            }
            #[cfg(not(target_os = "windows"))]
            {
                return Task::perform(async { "Not supported on this platform".to_string() }, Message::Error);
            }
        }
        _ => {}
    }

    // Update state with message
    message.update(&mut app.state)
}

/// View the application UI
pub fn view(app: &EnvManApp) -> Element<Message> {
    let content = if app.state.loading {
        column![
            iced::widget::text("Loading environment variables...")
                .size(18)
                .width(Length::Fill)
        ]
        .align_x(iced::Alignment::Center)
        .width(Length::Fill)
        .height(Length::Fill)
    } else {
        column![
            tool_bar(&app.state),
            tab_bar(&app.state),
            row![
                container(variable_list_panel(&app.state))
                    .width(Length::FillPortion(1))
                    .height(Length::Fill),
                container(editor_panel(&app.state))
                    .width(Length::FillPortion(2))
                    .height(Length::Fill),
            ]
            .height(Length::Fill)
            .spacing(8),
        ]
        .spacing(4)
    };

    // Overlay for messages
    let content_with_overlay = if let Some(ref error) = app.state.error_message {
        container(
            column![
                content,
                container(
                    column![
                        iced::widget::text("Error").size(16),
                        iced::widget::text(error).size(14),
                        iced::widget::button("Dismiss")
                            .on_press(Message::ClearMessages)
                    ]
                    .spacing(8)
                    .padding(16)
                    .width(Length::Fill)
                )
                .padding(16)
                .width(Length::Fill)
            ]
        )
        .width(Length::Fill)
        .height(Length::Fill)
    } else if let Some(ref success) = app.state.success_message {
        container(
            column![
                content,
                container(
                    column![
                        iced::widget::text("Success").size(16),
                        iced::widget::text(success).size(14),
                        iced::widget::button("Dismiss")
                            .on_press(Message::ClearMessages)
                    ]
                    .spacing(8)
                    .padding(16)
                    .width(Length::Fill)
                )
                .padding(16)
                .width(Length::Fill)
            ]
        )
        .width(Length::Fill)
        .height(Length::Fill)
    } else {
        container(content)
    };

    container(content_with_overlay)
        .padding(8)
        .width(Length::Fill)
        .height(Length::Fill)
        .into()
}
