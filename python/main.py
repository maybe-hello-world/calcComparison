from flask import Flask, request, make_response

app = Flask(__name__)


def collapse_tree(tree: dict) -> float:
    right = tree.get("right", None)
    if isinstance(right, dict):
        right = collapse_tree(right)
    if not (right and isinstance(right, (int, float))):
        raise ValueError("invalid data")

    left = tree.get("left", None)
    if isinstance(left, dict):
        left = collapse_tree(left)
    if not (left and isinstance(left, (int, float))):
        raise ValueError("invalid data")

    operation = tree.get("operation", None)
    if not (operation and isinstance(operation, str)):
        raise ValueError("invalid data")

    return do_operation(operation, left, right)


def do_operation(op: str, left: float, right: float) -> float:
    if op == "+":
        return left + right
    elif op == "-":
        return left - right
    elif op == "*":
        return left * right
    elif op == "/":
        if right == 0.0:
            raise ValueError("zero division error")
        return left / right
    else:
        raise ValueError("unsupported operation")


@app.route("/calc", methods=["POST"])
def calc_handler():
    data = request.get_json(silent=True)
    if not (data and isinstance(data, dict)):
        return make_response("invalid data", 400)

    try:
        return make_response({"result": collapse_tree(data)}, 200)
    except ValueError as e:
        return make_response(str(e), 400)


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8330)
