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

/// Technology represents a "technology" (in the very broad sense) that was used to create a
/// work.
#[derive(Serialize, Deserialize)]
pub struct Technology {
    /// Other technology slugs that refer to this technology. The slugs mentionned here should
    /// not be used in the definition of other technologies.
    pub aliases: Option<Vec<String>>,

    /// Autodetect contains an expression of the form 'CONTENT in PATH' where CONTENT is a
    /// free-form unquoted string and PATH is a filepath relative to the work folder.
    /// If CONTENT is found in PATH, we consider that technology to be used in the work.
    pub autodetect: Option<Vec<String>>,

    /// Name of the person or organization that created this technology.
    pub by: Option<String>,

    pub description: Option<String>,

    /// Files contains a list of gitignore-style patterns. If the work contains any of the
    /// patterns specified, we consider that technology to be used in the work.
    pub files: Option<Vec<String>>,

    /// URL to a website where more information can be found about this technology.
    #[serde(rename = "learn more at")]
    pub learn_more_at: Option<String>,

    pub name: String,

    /// The slug is a unique identifier for this technology, that's suitable for use in a
    /// website's URL.
    /// For example, the page that shows all works using a technology with slug "a" could be at
    /// https://example.org/technologies/a.
    pub slug: String,
}
