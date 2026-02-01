use super::variable::{EnvVariable, PathItem, VariableScope, VariableType};
use std::fmt;

/// Editor mode for the right panel
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum EditorMode {
    /// Plain text editing mode
    Text,
    /// List mode for PATH-like variables
    List,
}

impl fmt::Display for EditorMode {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            EditorMode::Text => write!(f, "Text"),
            EditorMode::List => write!(f, "List"),
        }
    }
}

/// Main application state
#[derive(Debug, Clone)]
pub struct AppState {
    /// System environment variables
    pub system_variables: Vec<EnvVariable>,
    /// User environment variables
    pub user_variables: Vec<EnvVariable>,
    /// Currently selected variable scope
    pub current_scope: VariableScope,
    /// Currently selected variable name (in the left panel)
    pub selected_variable: Option<String>,
    /// Editor mode
    pub editor_mode: EditorMode,
    /// Text being edited in the text editor
    pub editor_text: String,
    /// Path items being edited in list mode
    pub editor_path_items: Vec<PathItem>,
    /// Whether we're running as administrator
    pub is_elevated: bool,
    /// Error message to display
    pub error_message: Option<String>,
    /// Success message to display
    pub success_message: Option<String>,
    /// Whether there are unsaved changes
    pub has_unsaved_changes: bool,
    /// Currently loading data
    pub loading: bool,
}

impl Default for AppState {
    fn default() -> Self {
        Self {
            system_variables: Vec::new(),
            user_variables: Vec::new(),
            current_scope: VariableScope::User,
            selected_variable: None,
            editor_mode: EditorMode::Text,
            editor_text: String::new(),
            editor_path_items: Vec::new(),
            is_elevated: false,
            error_message: None,
            success_message: None,
            has_unsaved_changes: false,
            loading: true,
        }
    }
}

impl AppState {
    /// Create a new application state
    pub fn new() -> Self {
        Self::default()
    }

    /// Get the currently active variable list based on scope
    pub fn current_variables(&self) -> &[EnvVariable] {
        match self.current_scope {
            VariableScope::System => &self.system_variables,
            VariableScope::User => &self.system_variables,
        }
    }

    /// Get the currently active variable list as mutable
    pub fn current_variables_mut(&mut self) -> &mut Vec<EnvVariable> {
        match self.current_scope {
            VariableScope::System => &mut self.system_variables,
            VariableScope::User => &mut self.user_variables,
        }
    }

    /// Get the currently selected variable
    pub fn selected_variable_ref(&self) -> Option<&EnvVariable> {
        let vars = self.current_variables();
        vars.iter().find(|v| self.selected_variable.as_ref() == Some(&v.name))
    }

    /// Get the currently selected variable as mutable
    pub fn selected_variable_mut(&mut self) -> Option<&mut EnvVariable> {
        let selected = self.selected_variable.clone();
        let vars = self.current_variables_mut();
        vars.iter_mut().find(|v| Some(&v.name) == selected.as_ref())
    }

    /// Update variables for a specific scope
    pub fn set_variables(&mut self, scope: VariableScope, variables: Vec<EnvVariable>) {
        match scope {
            VariableScope::System => self.system_variables = variables,
            VariableScope::User => self.user_variables = variables,
        }
        self.loading = false;
    }

    /// Sort variables by name
    pub fn sort_variables(&mut self) {
        self.system_variables.sort_by(|a, b| a.name.cmp(&b.name));
        self.user_variables.sort_by(|a, b| a.name.cmp(&b.name));
    }

    /// Set an error message
    pub fn set_error(&mut self, error: String) {
        self.error_message = Some(error);
        self.success_message = None;
    }

    /// Set a success message
    pub fn set_success(&mut self, message: String) {
        self.success_message = Some(message);
        self.error_message = None;
    }

    /// Clear all messages
    pub fn clear_messages(&mut self) {
        self.error_message = None;
        self.success_message = None;
    }

    /// Update the editor content based on the selected variable
    pub fn update_editor_content(&mut self) {
        let value = self.selected_variable_ref().map(|v| v.value.clone());
        let var_type = self.selected_variable_ref().map(|v| v.variable_type);

        if let Some(value) = value {
            self.editor_text = value;

            if matches!(var_type, Some(VariableType::Path)) {
                let value_copy = self.editor_text.clone();
                self.editor_path_items = crate::services::parse_path_items(&value_copy);
                self.editor_mode = EditorMode::List;
            } else {
                self.editor_mode = EditorMode::Text;
            }
        } else {
            self.editor_text.clear();
            self.editor_path_items.clear();
        }
    }

    /// Check if there are any modified variables
    pub fn check_unsaved_changes(&self) -> bool {
        self.system_variables.iter().any(|v| v.modified)
            || self.user_variables.iter().any(|v| v.modified)
    }
}
