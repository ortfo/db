// Example code that deserializes and serializes the model.
// extern crate serde;
// #[macro_use]
// extern crate serde_derive;
// extern crate serde_json;
//
// use ortfodb::exporter;
//
// fn main() {
//     let json = r#"{"answer": 42}"#;
//     let model: exporter = serde_json::from_str(&json).unwrap();
// }

use serde::{Serialize, Deserialize};
use std::collections::HashMap;

#[derive(Serialize, Deserialize)]
pub struct Exporter {
    /// Commands to run after the build finishes. Go text template that receives .Data and
    /// .Database, the built database.
    pub after: Option<Vec<ExporterSchema>>,

    /// Commands to run before the build starts. Go text template that receives .Data
    pub before: Option<Vec<ExporterSchema>>,

    /// Initial data
    pub data: Option<HashMap<String, Option<serde_json::Value>>>,

    /// Some documentation about the exporter
    pub description: String,

    /// The name of the exporter
    pub name: String,

    /// List of programs that are required to be available in the PATH for the exporter to run.
    pub requires: Option<Vec<String>>,

    /// If true, will show every command that is run
    pub verbose: Option<bool>,

    /// Commands to run during the build, for each work. Go text template that receives .Data and
    /// .Work, the current work.
    pub work: Option<Vec<ExporterSchema>>,
}

#[derive(Serialize, Deserialize)]
pub struct ExporterSchema {
    /// Log a message. The first argument is the verb, the second is the color, the third is the
    /// message.
    pub log: Option<Vec<String>>,

    /// Run a command in a shell
    pub run: Option<String>,
}
