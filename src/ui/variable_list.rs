use iced::widget::{column, row, text, scrollable, container, button};
use iced::{Length, Alignment};

use crate::models::message::Message;
use crate::models::state::AppState;
use crate::models::variable::VariableScope;

/// Create the left panel showing the list of environment variables
pub fn variable_list_panel<'a>(state: &AppState) -> iced::Element<'a, Message> {
    let variables = match state.current_scope {
        VariableScope::System => &state.system_variables,
        VariableScope::User => &state.user_variables,
    };

    let header = row![
        text("Name").width(Length::FillPortion(1)),
        text("Value Preview").width(Length::FillPortion(2)),
        text("Type").width(Length::Fixed(60.0)),
    ]
    .spacing(8)
    .padding(8)
    .align_y(Alignment::Center);

    let list_content = if variables.is_empty() {
        column![text("No variables found").size(14)]
    } else {
        let items: Vec<iced::Element<Message>> = variables
            .iter()
            .map(|var| {
                let is_selected = state.selected_variable.as_ref() == Some(&var.name);

                let row_content = row![
                    {
                        let txt = if var.modified { format!("* {}", var.name) } else { var.name.clone() };
                        text(txt).width(Length::FillPortion(1))
                    },
                    text(var.value_preview(50))
                        .width(Length::FillPortion(2))
                        .size(12),
                    text(match var.variable_type {
                        crate::models::variable::VariableType::Path => "PATH",
                        crate::models::variable::VariableType::Standard => "Text",
                    })
                    .width(Length::Fixed(60.0))
                    .size(12),
                ]
                .spacing(8)
                .padding(8)
                .align_y(Alignment::Center);

                button(row_content)
                    .on_press(Message::SelectVariable(Some(var.name.clone())))
                    .width(Length::Fill)
                    .into()
            })
            .collect();

        column(items).spacing(2)
    };

    let content = column![header, scrollable(list_content).width(Length::Fill)]
        .spacing(4);

    container(content)
        .width(Length::Fill)
        .height(Length::Fill)
        .padding([8, 4])
        .into()
}
