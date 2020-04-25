use async_std::io;
use async_std::task;
use serde::{Deserialize, Serialize};
use tide::{Request, Response, StatusCode};

#[derive(Deserialize)]
#[serde(untagged)]
enum Node {
    Float(f64),
    Container(Container)
}

#[derive(Deserialize)]
struct Container {
    operation: String,
    left: Box<Node>,
    right: Box<Node>
}

#[derive(Serialize)]
struct CalcResult {
    result: f64
}

fn do_operation(op: String, left: f64, right: f64) -> Result<f64, &'static str> {
    match op.as_str() {
        "+" => Ok(left + right),
        "-" => Ok(left - right),
        "*" => Ok(left * right),
        "/" => if right == 0.0 { Err("zero division error") } else { Ok(left / right) }
         _  => Err("unsupported operation")
    }
}

fn collapse_tree(container: Container) -> Result<f64, &'static str> {
    let right = match *container.right {
        Node::Float(value) => value,
        Node::Container(subtree) => collapse_tree(subtree)?
    };

    let left = match *container.left {
        Node::Float(value) => value,
        Node::Container(subtree) => collapse_tree(subtree)?
    };

    do_operation(container.operation, left, right)
}

async fn calc_handler(mut req: Request<()>) -> tide::Result<Response> {
    let parse_result = req.body_json().await;

    let result = match parse_result {
        Ok(request) => collapse_tree(request),
        Err(some_error) => return Ok(Response::new(StatusCode::BadRequest).body_json(&some_error.to_string())?)
    };

    match result {
        Ok(calc_result) => Ok(Response::new(StatusCode::Ok).body_json(&CalcResult { result: calc_result })?),
        Err(err_text) => Ok(Response::new(StatusCode::BadRequest).body_json(&err_text)?)
    }
}

fn main() -> io::Result<()> {
    task::block_on(async {
        let mut app = tide::new();

        app.at("/calc").post(calc_handler);

        app.listen("0.0.0.0:8330").await?;
        Ok(())
    })
}