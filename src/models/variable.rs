use std::fmt;

/// Represents the type of an environment variable
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum VariableType {
    /// PATH-like variable with semicolon-separated values
    Path,
    /// Standard text variable
    Standard,
}

impl VariableType {
    /// Detect the variable type based on its name and value
    pub fn detect(name: &str, value: &str) -> Self {
        // Common PATH-like variables
        let path_vars = ["PATH", "Path", "path", "PATHEXT", "CLASSPATH", "PYTHONPATH"];

        if path_vars.contains(&name) || (value.contains(';') && value.contains(['\\', '/'])) {
            VariableType::Path
        } else {
            VariableType::Standard
        }
    }
}

/// Represents the scope of an environment variable
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum VariableScope {
    /// System-wide variables (requires admin to modify)
    System,
    /// Current user variables
    #[default]
    User,
}

impl fmt::Display for VariableScope {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            VariableScope::System => write!(f, "System"),
            VariableScope::User => write!(f, "User"),
        }
    }
}

/// Represents an environment variable with its metadata
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct EnvVariable {
    /// Variable name (key)
    pub name: String,
    /// Variable value
    pub value: String,
    /// Type of the variable
    pub variable_type: VariableType,
    /// Scope of the variable
    pub scope: VariableScope,
    /// Whether this variable has unsaved changes
    pub modified: bool,
}

impl EnvVariable {
    /// Create a new environment variable
    pub fn new(name: String, value: String, scope: VariableScope) -> Self {
        let variable_type = VariableType::detect(&name, &value);
        Self {
            name,
            value,
            variable_type,
            scope,
            modified: false,
        }
    }

    /// Create a new environment variable with explicit type
    pub fn with_type(name: String, value: String, scope: VariableScope, variable_type: VariableType) -> Self {
        Self {
            name,
            value,
            variable_type,
            scope,
            modified: false,
        }
    }

    /// Get a preview of the value (truncated if too long)
    pub fn value_preview(&self, max_len: usize) -> String {
        if self.value.len() <= max_len {
            self.value.clone()
        } else {
            format!("{}...", &self.value[..max_len.saturating_sub(3)])
        }
    }

    /// Mark the variable as modified/unmodified
    pub fn set_modified(&mut self, modified: bool) {
        self.modified = modified;
    }

    /// Update the value and mark as modified
    pub fn update_value(&mut self, new_value: String) {
        if self.value != new_value {
            self.value = new_value;
            self.modified = true;
        }
    }
}

/// Represents a single item in a PATH-like variable
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct PathItem {
    /// The path value
    pub value: String,
    /// Whether this item is valid (exists on disk)
    pub valid: bool,
    /// Whether this item is enabled (for future use with disabled items)
    pub enabled: bool,
}

impl PathItem {
    pub fn new(value: String) -> Self {
        Self {
            value,
            valid: false, // Will be validated asynchronously
            enabled: true,
        }
    }

    pub fn with_validation(value: String, valid: bool) -> Self {
        Self {
            value,
            valid,
            enabled: true,
        }
    }
}
