// Example code that deserializes and serializes the model.
// extern crate serde;
// #[macro_use]
// extern crate serde_derive;
// extern crate serde_json;
//
// use ortfodb::database;
//
// fn main() {
//     let json = r#"{"answer": 42}"#;
//     let model: database = serde_json::from_str(&json).unwrap();
// }

use serde::{Serialize, Deserialize};
use std::collections::HashMap;

pub type Database = HashMap<String, DatabaseValue>;

/// AnalyzedWork represents a complete work, with analyzed mediae.
#[derive(Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct DatabaseValue {
    pub built_at: String,

    pub content: HashMap<String, ContentValue>,

    pub description_hash: String,

    pub id: String,

    pub metadata: Metadata,

    #[serde(rename = "Partial")]
    pub partial: bool,
}

#[derive(Serialize, Deserialize)]
pub struct ContentValue {
    pub blocks: Vec<BlockElement>,

    pub footnotes: HashMap<String, String>,

    pub layout: Vec<Vec<String>>,

    pub title: String,
}

#[derive(Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct BlockElement {
    pub alt: String,

    /// whether the media has been analyzed
    pub analyzed: bool,

    pub anchor: String,

    pub attributes: Attributes,

    pub caption: String,

    pub colors: Colors,

    /// html
    pub content: String,

    pub content_type: String,

    pub dimensions: Dimensions,

    pub dist_source: String,

    /// in seconds
    pub duration: f64,

    pub has_sound: bool,

    pub id: String,

    pub index: i64,

    pub online: bool,

    pub relative_source: String,

    /// in bytes
    pub size: i64,

    pub text: String,

    pub thumbnails: Thumbnails,

    pub thumbnails_built_at: String,

    pub title: String,

    #[serde(rename = "type")]
    pub database_schema_type: String,

    pub url: String,
}

/// MediaAttributes stores which HTML attributes should be added to the media.
#[derive(Serialize, Deserialize)]
pub struct Attributes {
    /// Controlled with attribute character > (adds)
    pub autoplay: bool,

    /// Controlled with attribute character = (removes)
    pub controls: bool,

    /// Controlled with attribute character ~ (adds)
    #[serde(rename = "loop")]
    pub attributes_loop: bool,

    /// Controlled with attribute character > (adds)
    pub muted: bool,

    /// Controlled with attribute character = (adds)
    pub playsinline: bool,
}

/// ColorPalette reprensents the object in a Work's metadata.colors.
#[derive(Serialize, Deserialize)]
pub struct Colors {
    pub primary: String,

    pub secondary: String,

    pub tertiary: String,
}

/// ImageDimensions represents metadata about a media as it's extracted from its file.
#[derive(Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Dimensions {
    /// width / height
    pub aspect_ratio: f64,

    /// Height in pixels
    pub height: i64,

    /// Width in pixels
    pub width: i64,
}

#[derive(Serialize, Deserialize)]
pub struct Thumbnails {
}

#[derive(Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Metadata {
    pub additional_metadata: HashMap<String, Option<serde_json::Value>>,

    pub aliases: Vec<String>,

    pub colors: Colors,

    pub database_metadata: DatabaseMetadataClass,

    pub finished: String,

    pub made_with: Vec<String>,

    pub page_background: String,

    pub private: bool,

    pub started: String,

    pub tags: Vec<String>,

    pub thumbnail: String,

    pub title_style: String,

    pub wip: bool,
}

#[derive(Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct DatabaseMetadataClass {
    /// Partial is true if the database was not fully built.
    pub partial: bool,
}
