use iced::widget::{button, text, horizontal_space};
use iced::{Length, Alignment};

use crate::models::message::Message;
use crate::models::state::AppState;

/// Create a toolbar with action buttons
pub fn tool_bar<'a>(state: &AppState) -> iced::Element<'a, Message> {
    iced::widget::row![
        button("Refresh")
            .on_press(Message::Refresh),
        button("Save All")
            .on_press(Message::SaveAll),
        iced::widget::horizontal_space(),
        button(if state.is_elevated {
            "Restart without Admin"
        } else {
            "Restart as Admin"
        })
        .on_press(Message::RestartAsAdmin),
        text(if state.is_elevated {
            "Administrator"
        } else {
            "Standard User"
        }),
    ]
    .spacing(8)
    .padding(8)
    .align_y(Alignment::Center)
    .into()
}
