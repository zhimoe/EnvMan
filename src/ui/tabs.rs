use iced::widget::{button, horizontal_space};
use iced::{Length, Alignment};

use crate::models::message::Message;
use crate::models::state::AppState;
use crate::models::variable::VariableScope;

/// Create the tab bar for switching between System and User variables
pub fn tab_bar<'a>(_state: &AppState) -> iced::Element<'a, Message> {
    let system_tab = button("System Variables")
        .on_press(Message::SwitchScope(VariableScope::System));

    let user_tab = button("User Variables")
        .on_press(Message::SwitchScope(VariableScope::User));

    iced::widget::row![
        iced::widget::horizontal_space(),
        system_tab,
        user_tab,
        iced::widget::horizontal_space(),
    ]
    .spacing(8)
    .padding(8)
    .align_y(Alignment::Center)
    .width(Length::Fill)
    .into()
}
