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

/// Tag represents a category that can be assigned to a work.
#[derive(Serialize, Deserialize)]
pub struct Tag {
    /// Other singular-form names of tags that refer to this tag. The names mentionned here
    /// should not be used to define other tags.
    pub aliases: Option<Vec<String>>,

    pub description: Option<String>,

    /// Various ways to automatically detect that a work is tagged with this tag.
    pub detect: Option<Detect>,

    /// URL to a website where more information can be found about this tag.
    #[serde(rename = "learn more at")]
    pub learn_more_at: Option<String>,

    /// Plural-form name of the tag. For example, "Books".
    pub plural: String,

    /// Singular-form name of the tag. For example, "Book".
    pub singular: String,
}

/// Various ways to automatically detect that a work is tagged with this tag.
#[derive(Serialize, Deserialize)]
pub struct Detect {
    pub files: Option<Vec<String>>,

    #[serde(rename = "made with")]
    pub made_with: Option<Vec<String>>,

    pub search: Option<Vec<String>>,
}
