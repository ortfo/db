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

#[derive(Serialize, Deserialize)]
pub struct Database {
    #[serde(rename = "#meta")]
    pub meta: Option<Meta>,
}

#[derive(Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct Meta {
    pub partial: Option<bool>,
}
