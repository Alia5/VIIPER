use std::collections::HashMap;

pub use crate::devices::dualsense::DualsenseInput as DualsenseFullInput;
pub use crate::devices::dualsense::TriggerEffect;

pub mod output;
pub use output::*;

pub mod constants;
pub use constants::*;

pub use crate::devices::dualsense::meta::*;

#[derive(Debug, Clone, Default, serde::Serialize)]
pub struct CreateOptions {
    #[serde(rename = "Meta", skip_serializing_if = "Option::is_none")]
    pub meta: Option<MetaState>,
    #[serde(rename = "audio_port", skip_serializing_if = "Option::is_none")]
    pub audio_port: Option<u16>,
}

impl CreateOptions {
    pub fn to_map(&self) -> HashMap<String, serde_json::Value> {
        match serde_json::to_value(self) {
            Ok(serde_json::Value::Object(obj)) => obj.into_iter().collect(),
            _ => HashMap::new(),
        }
    }
}
