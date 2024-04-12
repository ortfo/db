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

    pub autodetect: Vec<String>,

    pub by: String,

    pub description: String,

    pub files: Vec<String>,

    #[serde(rename = "learn more at")]
    pub learn_more_at: String,

    pub name: String,

    pub slug: String,
}
