use crate::models::variable::{EnvVariable, VariableScope, VariableType};
use std::fmt;

/// Errors that can occur during registry operations
#[derive(Debug)]
pub enum RegistryError {
    AccessDenied(String),
    KeyNotFound(String),
    ValueNotFound(String),
    #[cfg(target_os = "windows")]
    WindowsError(windows::core::Error),
    IoError(std::io::Error),
    Other(String),
}

impl fmt::Display for RegistryError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            RegistryError::AccessDenied(s) => write!(f, "Registry access denied: {}", s),
            RegistryError::KeyNotFound(s) => write!(f, "Registry key not found: {}", s),
            RegistryError::ValueNotFound(s) => write!(f, "Value not found: {}", s),
            #[cfg(target_os = "windows")]
            RegistryError::WindowsError(e) => write!(f, "Windows API error: {}", e),
            RegistryError::IoError(e) => write!(f, "IO error: {}", e),
            RegistryError::Other(s) => write!(f, "Other error: {}", s),
        }
    }
}

impl std::error::Error for RegistryError {}

impl From<std::io::Error> for RegistryError {
    fn from(e: std::io::Error) -> Self {
        RegistryError::IoError(e)
    }
}

#[cfg(target_os = "windows")]
impl From<windows::core::Error> for RegistryError {
    fn from(e: windows::core::Error) -> Self {
        RegistryError::WindowsError(e)
    }
}

/// Registry key paths for environment variables
const SYSTEM_ENV_KEY: &str = r"SYSTEM\CurrentControlSet\Control\Session Manager\Environment";
const USER_ENV_KEY: &str = r"Environment";

/// Get the registry key path for a given scope
fn get_registry_key_path(scope: VariableScope) -> &'static str {
    match scope {
        VariableScope::System => SYSTEM_ENV_KEY,
        VariableScope::User => USER_ENV_KEY,
    }
}

/// Load environment variables from the registry
#[cfg(target_os = "windows")]
pub async fn load_variables(scope: VariableScope) -> Result<Vec<EnvVariable>, RegistryError> {
    use winreg::enums::*;
    use winreg::RegKey;

    let hkey = match scope {
        VariableScope::System => HKEY_LOCAL_MACHINE,
        VariableScope::User => HKEY_CURRENT_USER,
    };

    let key_path = get_registry_key_path(scope);

    // Open the registry key
    let env_key = RegKey::predef(hkey)
        .open_subkey(key_path)
        .map_err(|e| RegistryError::AccessDenied(format!("Cannot open {}: {}", key_path, e)))?;

    // Enumerate all values
    let mut variables = Vec::new();

    for (name, value) in env_key.enum_values().flatten() {
        let value_str = match value {
            winreg::RegValue::Sz(s) | winreg::RegValue::ExpandSz(s) => s,
            // Handle DWORD values as strings (some env vars like NUMBER_OF_PROCESSORS)
            winreg::RegValue::Dword(d) => d.to_string(),
            winreg::RegValue::Qword(q) => q.to_string(),
            _ => continue,
        };

        // Expand environment variables in the value
        let expanded_value = expand_environment_vars(&value_str);

        let variable_type = VariableType::detect(&name, &expanded_value);

        variables.push(EnvVariable {
            name,
            value: expanded_value,
            variable_type,
            scope,
            modified: false,
        });
    }

    // Sort by name
    variables.sort_by(|a, b| a.name.cmp(&b.name));

    Ok(variables)
}

/// Load environment variables from the registry (non-Windows stub)
#[cfg(not(target_os = "windows"))]
pub async fn load_variables(scope: VariableScope) -> Result<Vec<EnvVariable>, RegistryError> {
    // Return some dummy variables for testing on non-Windows platforms
    let mut variables = Vec::new();

    if matches!(scope, VariableScope::User) {
        variables.push(EnvVariable::new(
            "PATH".to_string(),
            "/usr/bin:/usr/local/bin".to_string(),
            scope,
        ));
        variables.push(EnvVariable::new(
            "HOME".to_string(),
            "/home/user".to_string(),
            scope,
        ));
    }

    Ok(variables)
}

/// Save environment variables to the registry
#[cfg(target_os = "windows")]
pub async fn save_variables(
    scope: VariableScope,
    variables: &[EnvVariable],
) -> Result<(), RegistryError> {
    use winreg::enums::*;
    use winreg::RegKey;

    let hkey = match scope {
        VariableScope::System => HKEY_LOCAL_MACHINE,
        VariableScope::User => HKEY_CURRENT_USER,
    };

    let key_path = get_registry_key_path(scope);

    // Open the registry key with write access
    let mut env_key = RegKey::predef(hkey)
        .open_subkey_with_flags(key_path, KEY_SET_VALUE | KEY_READ)
        .map_err(|e| RegistryError::AccessDenied(format!("Cannot open {} for writing: {}", key_path, e)))?;

    // Get existing variables to detect deletions
    let existing: HashMap<String, bool> = env_key
        .enum_values()
        .flatten()
        .map(|(name, _)| (name, true))
        .collect();

    // Write all variables
    for var in variables {
        if var.name.is_empty() {
            continue;
        }

        // Determine if we need to use REG_EXPAND_SZ
        // Use EXPAND_SZ if the value contains environment variable references
        let needs_expand = var.value.contains('%') || var.value.contains('$');

        if needs_expand {
            env_key.set_value(&var.name, &var.value.as_str())
                .map_err(|e| RegistryError::Other(format!("Failed to set {}: {}", var.name, e)))?;
        } else {
            env_key.set_value(&var.name, &var.value.as_str())
                .map_err(|e| RegistryError::Other(format!("Failed to set {}: {}", var.name, e)))?;
        }
    }

    // Broadcast the change notification
    broadcast_environment_change();

    Ok(())
}

/// Save environment variables (non-Windows stub)
#[cfg(not(target_os = "windows"))]
pub async fn save_variables(
    _scope: VariableScope,
    _variables: &[EnvVariable],
) -> Result<(), RegistryError> {
    // Stub for non-Windows platforms
    Ok(())
}

/// Check if the current process is running with administrator privileges
#[cfg(target_os = "windows")]
pub async fn is_elevated() -> Result<bool, RegistryError> {
    use windows::Win32::Foundation::HANDLE;
    use windows::Win32::System::Threading::{GetCurrentProcess, IsUserAnAdmin};

    unsafe {
        Ok(IsUserAnAdmin().as_bool())
    }
}

/// Check elevation status (non-Windows stub)
#[cfg(not(target_os = "windows"))]
pub async fn is_elevated() -> Result<bool, RegistryError> {
    Ok(false)
}

/// Restart the application with administrator privileges
#[cfg(target_os = "windows")]
pub fn restart_as_admin() -> Result<(), RegistryError> {
    use std::env;
    use std::ffi::OsString;
    use windows::core::PCWSTR;
    use windows::Win32::Foundation::HWND;
    use windows::Win32::UI::Shell::ShellExecuteW;

    let current_exe = env::current_exe()
        .map_err(|e| RegistryError::Other(format!("Cannot get current executable: {}", e)))?;

    let exe_path: Vec<u16> = current_exe
        .as_os_str()
        .encode_wide()
        .chain(std::iter::once(0))
        .collect();

    unsafe {
        ShellExecuteW(
            HWND::default(),
            w!("runas"),
            PCWSTR(exe_path.as_ptr()),
            PCWSTR::null(),
            PCWSTR::null(),
            1, // SW_SHOWNORMAL
        );
    }

    Ok(())
}

/// Restart as admin (non-Windows stub)
#[cfg(not(target_os = "windows"))]
pub fn restart_as_admin() -> Result<(), RegistryError> {
    Err(RegistryError::Other("Administrator elevation not supported on this platform".to_string()))
}

/// Broadcast WM_SETTINGCHANGE to notify other applications of environment changes
#[cfg(target_os = "windows")]
fn broadcast_environment_change() {
    use windows::Win32::Foundation::HWND;
    use windows::Win32::UI::WindowsAndMessaging::{
        SendMessageTimeoutW, HWND_BROADCAST, WM_SETTINGCHANGE, SMTO_ABORTIFHUNG, LPARAM, LRESULT, WPARAM,
    };
    use windows::core::PCWSTR;

    let environment: Vec<u16> = "Environment"
        .encode_utf16()
        .chain(std::iter::once(0))
        .collect();

    unsafe {
        SendMessageTimeoutW(
            HWND_BROADCAST,
            WM_SETTINGCHANGE,
            WPARAM(0),
            LPARAM(environment.as_ptr() as isize),
            SMTO_ABORTIFHUNG,
            5000,
            None,
        );
    }
}

/// Expand environment variables in a string
#[cfg(target_os = "windows")]
fn expand_environment_vars(value: &str) -> String {
    use windows::Win32::System::Registry::ExpandEnvironmentStringsW;

    if !value.contains('%') {
        return value.to_string();
    }

    let input: Vec<u16> = value.encode_utf16().chain(std::iter::once(0)).collect();

    unsafe {
        // First call to get the required buffer size
        let size = ExpandEnvironmentStringsW(PCWSTR(input.as_ptr()), PCWSTR::null(), 0);

        if size == 0 {
            return value.to_string();
        }

        let mut buffer = vec![0u16; size as usize];
        let result = ExpandEnvironmentStringsW(
            PCWSTR(input.as_ptr()),
            PCWSTR(buffer.as_mut_ptr()),
            size,
        );

        if result == 0 {
            value.to_string()
        } else {
            String::from_utf16_lossy(&buffer[..(result as usize - 1)])
        }
    }
}

/// Expand environment variables (non-Windows stub)
#[cfg(not(target_os = "windows"))]
fn expand_environment_vars(value: &str) -> String {
    value.to_string()
}
