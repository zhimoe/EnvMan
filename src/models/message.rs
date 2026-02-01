use super::variable::VariableScope;
use super::state::EditorMode;
use super::state::AppState;

/// Messages for the Elm-style architecture
#[derive(Debug, Clone)]
pub enum Message {
    // ==================== Application Lifecycle ====================
    /// Initialize the application, load variables
    Initialize,
    /// Application loaded successfully
    Loaded,
    /// Error occurred
    Error(String),

    // ==================== Menu Bar Actions ====================
    /// Restart with administrator privileges
    RestartAsAdmin,
    /// Refresh variable list from registry
    Refresh,
    /// Exit application
    Exit,

    // ==================== Tab Actions ====================
    /// Switch between System and User variables
    SwitchScope(VariableScope),

    // ==================== Variable List Actions ====================
    /// Select a variable in the list
    SelectVariable(Option<String>),
    /// Add a new variable
    AddVariable,
    /// Delete the selected variable
    DeleteVariable,
    /// Duplicate the selected variable
    DuplicateVariable,

    // ==================== Editor Actions ====================
    /// Switch editor mode (Text/List)
    SwitchEditorMode(EditorMode),
    /// Text editor content changed
    EditorTextChanged(String),
    /// Save changes from editor to the selected variable
    SaveEditorChanges,
    /// Discard changes to the selected variable
    DiscardEditorChanges,

    // ==================== List Editor Actions ====================
    /// Add a new item to the path list
    AddPathItem,
    /// Remove an item from the path list
    RemovePathItem(usize),
    /// Move a path item up
    MovePathItemUp(usize),
    /// Move a path item down
    MovePathItemDown(usize),
    /// Edit a path item value
    EditPathItem(usize, String),
    /// Validate all path items (check if they exist)
    ValidatePathItems,
    /// Toggle a path item enabled/disabled
    TogglePathItem(usize),

    // ==================== Save/Load Actions ====================
    /// Save all changes to registry
    SaveAll,
    /// Reload from registry (discarding changes)
    Reload,

    // ==================== Status Messages ====================
    /// Clear error/success messages
    ClearMessages,
    /// Show a toast/notification (for future use)
    ShowNotification(String),

    // ==================== Search/Filter ====================
    /// Filter the variable list
    FilterVariables(String),

    // ==================== Async Operations ====================
    /// Variables loaded from registry
    VariablesLoaded(VariableScope, Vec<crate::models::variable::EnvVariable>),
    /// Changes saved to registry
    ChangesSaved(VariableScope),
    /// Admin status checked
    AdminStatusChecked(bool),
}

impl Message {
    /// Update the application state based on this message
    pub fn update(self, state: &mut AppState) -> iced::Task<Message> {
        match self {
            Message::Initialize => {
                state.loading = true;
                // Trigger loading of both system and user variables
                iced::Task::batch([
                    iced::Task::perform(load_system_variables(), |vars| {
                        Message::VariablesLoaded(VariableScope::System, vars)
                    }),
                    iced::Task::perform(load_user_variables(), |vars| {
                        Message::VariablesLoaded(VariableScope::User, vars)
                    }),
                    iced::Task::perform(check_admin_status(), |is_admin| {
                        Message::AdminStatusChecked(is_admin)
                    }),
                ])
            }

            Message::VariablesLoaded(scope, vars) => {
                state.set_variables(scope, vars);
                state.sort_variables();
                if state.current_scope == scope {
                    state.update_editor_content();
                }
                iced::Task::none()
            }

            Message::AdminStatusChecked(is_admin) => {
                state.is_elevated = is_admin;
                iced::Task::none()
            }

            Message::SwitchScope(scope) => {
                state.current_scope = scope;
                state.selected_variable = None;
                state.update_editor_content();
                iced::Task::none()
            }

            Message::SelectVariable(name) => {
                state.selected_variable = name;
                state.update_editor_content();
                iced::Task::none()
            }

            Message::SwitchEditorMode(mode) => {
                state.editor_mode = mode;
                if mode == EditorMode::List {
                    if let Some(var) = state.selected_variable_ref() {
                        state.editor_path_items = crate::services::parse_path_items(&var.value);
                    }
                }
                iced::Task::none()
            }

            Message::EditorTextChanged(text) => {
                state.editor_text = text;
                iced::Task::none()
            }

            Message::SaveEditorChanges => {
                let editor_mode = state.editor_mode;
                let editor_text = state.editor_text.clone();
                let editor_path_items = state.editor_path_items.clone();

                if let Some(var) = state.selected_variable_mut() {
                    let new_value = if editor_mode == EditorMode::Text {
                        editor_text
                    } else {
                        crate::services::join_path_items(&editor_path_items)
                    };

                    if var.value != new_value {
                        var.update_value(new_value);
                        state.has_unsaved_changes = true;
                    }
                }
                iced::Task::none()
            }

            Message::DiscardEditorChanges => {
                state.update_editor_content();
                iced::Task::none()
            }

            Message::AddPathItem => {
                state.editor_path_items.push(crate::models::variable::PathItem::new(String::new()));
                iced::Task::none()
            }

            Message::RemovePathItem(index) => {
                if index < state.editor_path_items.len() {
                    state.editor_path_items.remove(index);
                }
                iced::Task::none()
            }

            Message::MovePathItemUp(index) => {
                if index > 0 && index < state.editor_path_items.len() {
                    state.editor_path_items.swap(index, index - 1);
                }
                iced::Task::none()
            }

            Message::MovePathItemDown(index) => {
                if index + 1 < state.editor_path_items.len() {
                    state.editor_path_items.swap(index, index + 1);
                }
                iced::Task::none()
            }

            Message::EditPathItem(index, value) => {
                if index < state.editor_path_items.len() {
                    state.editor_path_items[index].value = value;
                }
                iced::Task::none()
            }

            Message::TogglePathItem(index) => {
                if index < state.editor_path_items.len() {
                    state.editor_path_items[index].enabled = !state.editor_path_items[index].enabled;
                }
                iced::Task::none()
            }

            Message::ValidatePathItems => {
                iced::Task::perform(validate_paths(state.editor_path_items.clone()), |_items| {
                    // Update state with validated items
                    Message::ShowNotification("Paths validated".to_string())
                })
            }

            Message::SaveAll => {
                state.loading = true;
                let system_vars = state.system_variables.clone();
                let user_vars = state.user_variables.clone();

                iced::Task::batch([
                    iced::Task::perform(
                        save_variables(VariableScope::System, system_vars),
                        |_| Message::ChangesSaved(VariableScope::System)
                    ),
                    iced::Task::perform(
                        save_variables(VariableScope::User, user_vars),
                        |_| Message::ChangesSaved(VariableScope::User)
                    ),
                ])
            }

            Message::ChangesSaved(_) => {
                state.loading = false;
                state.has_unsaved_changes = false;
                // Clear modified flags
                for var in &mut state.system_variables {
                    var.set_modified(false);
                }
                for var in &mut state.user_variables {
                    var.set_modified(false);
                }
                state.set_success("Changes saved successfully".to_string());
                iced::Task::none()
            }

            Message::Reload => {
                state.loading = true;
                state.clear_messages();
                iced::Task::batch([
                    iced::Task::perform(load_system_variables(), |vars| {
                        Message::VariablesLoaded(VariableScope::System, vars)
                    }),
                    iced::Task::perform(load_user_variables(), |vars| {
                        Message::VariablesLoaded(VariableScope::User, vars)
                    }),
                ])
            }

            Message::ClearMessages => {
                state.clear_messages();
                iced::Task::none()
            }

            Message::Error(msg) => {
                state.set_error(msg);
                state.loading = false;
                iced::Task::none()
            }

            Message::RestartAsAdmin => {
                // Handle in the UI layer
                iced::Task::none()
            }

            Message::Refresh => {
                state.loading = true;
                iced::Task::batch([
                    iced::Task::perform(load_system_variables(), |vars| {
                        Message::VariablesLoaded(VariableScope::System, vars)
                    }),
                    iced::Task::perform(load_user_variables(), |vars| {
                        Message::VariablesLoaded(VariableScope::User, vars)
                    }),
                ])
            }

            Message::FilterVariables(_filter) => {
                // Filter implementation would go here
                iced::Task::none()
            }

            Message::AddVariable => {
                // Add new variable implementation
                iced::Task::none()
            }

            Message::DeleteVariable => {
                // Delete variable implementation
                iced::Task::none()
            }

            Message::DuplicateVariable => {
                // Duplicate variable implementation
                iced::Task::none()
            }

            Message::Exit => {
                iced::Task::none()
            }

            Message::Loaded | Message::ShowNotification(_) => {
                iced::Task::none()
            }
        }
    }
}

// Async function stubs - implementations will be in the services module
async fn load_system_variables() -> Vec<crate::models::variable::EnvVariable> {
    crate::services::load_variables(VariableScope::System).await.unwrap_or_default()
}

async fn load_user_variables() -> Vec<crate::models::variable::EnvVariable> {
    crate::services::load_variables(VariableScope::User).await.unwrap_or_default()
}

async fn check_admin_status() -> bool {
    crate::services::is_elevated().await.unwrap_or(false)
}

async fn save_variables(
    scope: VariableScope,
    variables: Vec<crate::models::variable::EnvVariable>,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    crate::services::save_variables(scope, &variables).await
        .map_err(|e| Box::new(e) as Box<dyn std::error::Error + Send + Sync>)
}

async fn validate_paths(items: Vec<crate::models::variable::PathItem>) -> Vec<crate::models::variable::PathItem> {
    // Validate paths - implementation in services
    items
}
