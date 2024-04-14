// Example code that deserializes and serializes the model.
// extern crate serde;
// #[macro_use]
// extern crate serde_derive;
// extern crate serde_json;
//
// use ortfodb::technologies;
//
// fn main() {
//     let json = r#"{"answer": 42}"#;
//     let model: technologies = serde_json::from_str(&json).unwrap();
// }

use serde::{Serialize, Deserialize};

pub type Technologies = Vec<Technology>;

#[derive(Serialize, Deserialize)]
pub struct Technology {
    pub aliases: Vec<String>,

    /// Autodetect contains an expression of the form 'CONTENT in PATH' where CONTENT is a
    /// free-form unquoted string and PATH is a filepath relative to the work folder.
    /// If CONTENT is found in PATH, we consider that technology to be used in the work.
    pub autodetect: Vec<String>,

    pub by: String,

    pub description: String,

    /// Files contains a list of gitignore-style patterns. If the work contains any of the
    /// patterns specified, we consider that technology to be used in the work.
    pub files: Vec<String>,

    #[serde(rename = "learn more at")]
    pub learn_more_at: String,

    pub name: String,

    pub slug: String,
}
