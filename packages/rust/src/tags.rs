// Example code that deserializes and serializes the model.
// extern crate serde;
// #[macro_use]
// extern crate serde_derive;
// extern crate serde_json;
//
// use ortfodb::tags;
//
// fn main() {
//     let json = r#"{"answer": 42}"#;
//     let model: tags = serde_json::from_str(&json).unwrap();
// }

use serde::{Serialize, Deserialize};

pub type Tags = Vec<Tag>;

#[derive(Serialize, Deserialize)]
pub struct Tag {
    pub aliases: Vec<String>,

    pub description: String,

    pub detect: Detect,

    #[serde(rename = "learn more at")]
    pub learn_more_at: String,

    pub plural: String,

    pub singular: String,
}

#[derive(Serialize, Deserialize)]
pub struct Detect {
    pub files: Vec<String>,

    #[serde(rename = "made with")]
    pub made_with: Vec<String>,

    pub search: Vec<String>,
}
