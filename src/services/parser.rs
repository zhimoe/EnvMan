use crate::models::variable::PathItem;

/// Parse a PATH-like string into individual path items
pub fn parse_path_items(value: &str) -> Vec<PathItem> {
    if value.is_empty() {
        return Vec::new();
    }

    value
        .split(';')
        .map(|s| s.trim())
        .filter(|s| !s.is_empty())
        .map(|s| PathItem::new(s.to_string()))
        .collect()
}

/// Join path items back into a semicolon-separated string
pub fn join_path_items(items: &[PathItem]) -> String {
    items
        .iter()
        .filter(|item| item.enabled && !item.value.is_empty())
        .map(|item| item.value.as_str())
        .collect::<Vec<_>>()
        .join(";")
}

/// Validate if a path exists on the filesystem
pub fn validate_path(path: &str) -> bool {
    if path.is_empty() {
        return false;
    }

    #[cfg(target_os = "windows")]
    let path_expanded = expand_path(path);
    #[cfg(not(target_os = "windows"))]
    let path_expanded = path;

    std::path::Path::new(path_expanded).exists()
}

/// Expand environment variables in a path (e.g., %SystemRoot%)
#[cfg(target_os = "windows")]
fn expand_path(path: &str) -> String {
    use windows::Win32::System::Registry::ExpandEnvironmentStringsW;
    use windows::core::PCWSTR;

    if !path.contains('%') {
        return path.to_string();
    }

    let input: Vec<u16> = path.encode_utf16().chain(std::iter::once(0)).collect();

    unsafe {
        let size = ExpandEnvironmentStringsW(PCWSTR(input.as_ptr()), PCWSTR::null(), 0);

        if size == 0 {
            return path.to_string();
        }

        let mut buffer = vec![0u16; size as usize];
        let result = ExpandEnvironmentStringsW(
            PCWSTR(input.as_ptr()),
            PCWSTR(buffer.as_mut_ptr()),
            size,
        );

        if result == 0 {
            path.to_string()
        } else {
            String::from_utf16_lossy(&buffer[..(result as usize - 1)])
        }
    }
}

/// Expand path (non-Windows stub)
#[cfg(not(target_os = "windows"))]
fn expand_path(path: &str) -> String {
    path.to_string()
}

/// Clean up a path string (normalize separators, trim whitespace)
pub fn clean_path(path: &str) -> String {
    let path = path.trim();

    #[cfg(target_os = "windows")]
    {
        // Normalize backslashes to forward slashes or vice versa
        // and remove trailing slashes (except for root paths)
        let normalized = path.replace('/', "\\");
        if normalized.len() > 1 && (normalized.ends_with('\\') || normalized.ends_with('/')) {
            let trimmed = &normalized[..normalized.len() - 1];
            if !trimmed.ends_with(':') {
                return trimmed.to_string();
            }
        }
        normalized
    }

    #[cfg(not(target_os = "windows"))]
    {
        path.to_string()
    }
}

/// Check if a path is a valid UNC path (Windows)
#[cfg(target_os = "windows")]
pub fn is_unc_path(path: &str) -> bool {
    path.starts_with("\\\\") || path.starts_with("//")
}

/// Check if a path is a valid UNC path (non-Windows stub)
#[cfg(not(target_os = "windows"))]
pub fn is_unc_path(_path: &str) -> bool {
    false
}

/// Extract the drive letter from a Windows path (e.g., "C:" from "C:\Windows")
#[cfg(target_os = "windows")]
pub fn extract_drive_letter(path: &str) -> Option<char> {
    let chars: Vec<char> = path.chars().collect();
    if chars.len() >= 2 && chars[0].is_ascii_alphabetic() && chars[1] == ':' {
        Some(chars[0])
    } else {
        None
    }
}

/// Extract the drive letter (non-Windows stub)
#[cfg(not(target_os = "windows"))]
pub fn extract_drive_letter(_path: &str) -> Option<char> {
    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_path_items() {
        let path = r"C:\Windows;C:\Program Files;D:\Tools";
        let items = parse_path_items(path);
        assert_eq!(items.len(), 3);
        assert_eq!(items[0].value, r"C:\Windows");
        assert_eq!(items[1].value, r"C:\Program Files");
        assert_eq!(items[2].value, r"D:\Tools");
    }

    #[test]
    fn test_parse_empty_path() {
        let items = parse_path_items("");
        assert_eq!(items.len(), 0);
    }

    #[test]
    fn test_join_path_items() {
        let items = vec![
            PathItem::new(r"C:\Windows".to_string()),
            PathItem::new(r"C:\Program Files".to_string()),
        ];
        let joined = join_path_items(&items);
        assert_eq!(joined, r"C:\Windows;C:\Program Files");
    }

    #[test]
    fn test_join_path_items_with_disabled() {
        let mut items = vec![
            PathItem::new(r"C:\Windows".to_string()),
            PathItem::new(r"C:\Disabled".to_string()),
        ];
        items[1].enabled = false;
        let joined = join_path_items(&items);
        assert_eq!(joined, r"C:\Windows");
    }

    #[test]
    fn test_clean_path() {
        assert_eq!(clean_path(r" C:\Windows\ "), r"C:\Windows");
        assert_eq!(clean_path(r"C:/Windows/System32"), r"C:\Windows\System32");
    }
}
