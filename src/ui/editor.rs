use iced::widget::{column, row, text, text_input, button, scrollable, container, pick_list};
use iced::{Length, Alignment};

use crate::models::message::Message;
use crate::models::state::{AppState, EditorMode};
use crate::models::variable::VariableType;

/// Editor mode options
const EDITOR_MODES: [EditorMode; 2] = [EditorMode::Text, EditorMode::List];

/// Create the right panel editor for the selected variable
pub fn editor_panel<'a>(state: &AppState) -> iced::Element<'a, Message> {
    if let Some(var) = state.selected_variable_ref() {
        column![
            editor_header(state, var),
            editor_content(state, var),
            editor_footer(state),
        ]
        .spacing(8)
        .padding(8)
        .width(Length::Fill)
        .height(Length::Fill)
        .into()
    } else {
        container(
            column![
                text("No variable selected")
                    .size(16)
                    .width(Length::Fill),
                text("Select a variable from the list to edit its value")
                    .size(14)
                    .width(Length::Fill),
            ]
            .align_x(Alignment::Center)
            .width(Length::Fill)
        )
        .width(Length::Fill)
        .height(Length::Fill)
        .padding(16)
        .center_x(Length::Fill)
        .center_y(Length::Fill)
        .into()
    }
}

/// Editor header showing variable name and mode selector
fn editor_header<'a>(state: &AppState, var: &crate::models::variable::EnvVariable) -> iced::Element<'a, Message> {
    row![
        text(format!("Name: {}", var.name)).size(14),
        text(format!("Type: {:?}", var.variable_type)).size(12),
        iced::widget::horizontal_space(),
        text("Mode:"),
        pick_list(
            EDITOR_MODES,
            Some(state.editor_mode),
            Message::SwitchEditorMode
        ),
    ]
    .spacing(8)
    .padding(8)
    .align_y(Alignment::Center)
    .into()
}

/// Editor content based on current mode
fn editor_content<'a>(state: &AppState, var: &crate::models::variable::EnvVariable) -> iced::Element<'a, Message> {
    match state.editor_mode {
        EditorMode::Text => text_editor(state),
        EditorMode::List => {
            if var.variable_type == VariableType::Path {
                path_list_editor(state)
            } else {
                text("This variable is not a PATH-type variable. Use Text mode to edit.")
                    .size(12)
                    .width(Length::Fill)
                    .into()
            }
        }
    }
}

/// Text mode editor
fn text_editor<'a>(state: &AppState) -> iced::Element<'a, Message> {
    scrollable(
        text_input("Enter value...", &state.editor_text)
            .on_input(Message::EditorTextChanged)
            .padding(8)
            .width(Length::Fill)
    )
    .width(Length::Fill)
    .height(Length::Fill)
    .into()
}

/// List mode editor for PATH-like variables
fn path_list_editor<'a>(state: &AppState) -> iced::Element<'a, Message> {
    if state.editor_path_items.is_empty() {
        return column![
            text("No path items").size(14),
            button("Add Item").on_press(Message::AddPathItem),
        ]
        .spacing(8)
        .align_x(Alignment::Center)
        .width(Length::Fill)
        .into();
    }

    let items: Vec<iced::Element<Message>> = state
        .editor_path_items
        .iter()
        .enumerate()
        .map(|(index, item)| {
            row![
                button("▲")
                    .on_press(Message::MovePathItemUp(index))
                    .width(Length::Fixed(30.0)),
                button("▼")
                    .on_press(Message::MovePathItemDown(index))
                    .width(Length::Fixed(30.0)),
                button("✕")
                    .on_press(Message::RemovePathItem(index))
                    .width(Length::Fixed(30.0)),
                text_input("Path...", &item.value)
                    .on_input(move |s| Message::EditPathItem(index, s))
                    .width(Length::Fill),
                button(if item.enabled { "☑" } else { "☐" })
                    .on_press(Message::TogglePathItem(index))
                    .width(Length::Fixed(40.0)),
                {
                    let valid = item.valid;
                    text(if valid { "✓" } else { "!" })
                        .width(Length::Fixed(20.0))
                },
            ]
            .spacing(4)
            .align_y(Alignment::Center)
            .into()
        })
        .collect();

    scrollable(
        column(items)
            .spacing(4)
            .padding(8)
            .push(button("Add Item").on_press(Message::AddPathItem))
    )
    .width(Length::Fill)
    .height(Length::Fill)
    .into()
}

/// Editor footer with action buttons
fn editor_footer<'a>(_state: &AppState) -> iced::Element<'a, Message> {
    row![
        button("Save").on_press(Message::SaveEditorChanges),
        button("Discard").on_press(Message::DiscardEditorChanges),
    ]
    .spacing(8)
    .align_y(Alignment::Center)
    .into()
}
