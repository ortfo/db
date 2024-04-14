// Example code that deserializes and serializes the model.
// extern crate serde;
// #[macro_use]
// extern crate serde_derive;
// extern crate serde_json;
//
// use ortfodb::configuration;
//
// fn main() {
//     let json = r#"{"answer": 42}"#;
//     let model: configuration = serde_json::from_str(&json).unwrap();
// }

use serde::{Serialize, Deserialize};
use std::collections::HashMap;

/// Configuration represents what the ortfodb.yaml configuration file describes.
#[derive(Serialize, Deserialize)]
pub struct Configuration {
    #[serde(rename = "build metadata file")]
    pub build_metadata_file: String,

    /// Exporter-specific configuration. Maps exporter names to their configuration.
    pub exporters: Option<HashMap<String, HashMap<String, Option<serde_json::Value>>>>,

    #[serde(rename = "extract colors")]
    pub extract_colors: ExtractColors,

    #[serde(rename = "make gifs")]
    pub make_gifs: MakeGifs,

    #[serde(rename = "make thumbnails")]
    pub make_thumbnails: MakeThumbnails,

    pub media: Media,

    /// Path to the directory containing all projects. Must be absolute.
    #[serde(rename = "projects at")]
    pub projects_at: String,

    #[serde(rename = "scattered mode folder")]
    pub scattered_mode_folder: String,

    pub tags: Tags,

    pub technologies: Technologies,
}

#[derive(Serialize, Deserialize)]
pub struct ExtractColors {
    #[serde(rename = "default files")]
    pub default_files: Vec<String>,

    pub enabled: bool,

    pub extract: Vec<String>,
}

#[derive(Serialize, Deserialize)]
pub struct MakeGifs {
    pub enabled: bool,

    #[serde(rename = "file name template")]
    pub file_name_template: String,
}

#[derive(Serialize, Deserialize)]
pub struct MakeThumbnails {
    pub enabled: bool,

    #[serde(rename = "file name template")]
    pub file_name_template: String,

    #[serde(rename = "input file")]
    pub input_file: String,

    pub sizes: Vec<i64>,
}

#[derive(Serialize, Deserialize)]
pub struct Media {
    pub at: String,
}

#[derive(Serialize, Deserialize)]
pub struct Tags {
    pub repository: String,
}

#[derive(Serialize, Deserialize)]
pub struct Technologies {
    pub repository: String,
}
