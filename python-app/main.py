from flask import Flask, request, jsonify, Response
import logging
import os
from upscaler import Upscaler

app = Flask(__name__)
UPLOAD_DIR = "./uploads/"
upscaler = Upscaler('models/RealESRGAN_x4plus.pth', 4)
upscaler.new_default()
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

@app.route('/upscale', methods=["POST"])
def upscale_file() -> tuple[Response, int]:
    # get filename from data
    data = request.get_json()
    filename: str = data.get('filename')

    # check if filename exists
    if filename == "":
        return jsonify({
            "status": 404,
            "message": "no file provided",
            "file": "",
        }), 404

    # upscale image
    try:
        file_path: str = os.path.join(os.getcwd(), "uploads", filename)

        if not os.path.exists(file_path):
            return jsonify({
                "status": 404,
                "message": f"file {filename} not found",
                "file": filename,
            }), 404

        result = upscaler.process(file_path)
        if isinstance(result, Exception):
            logging.info("failed upscaling")
            return jsonify({
                "status": "500",
                "message": f"encountered error: {str(result)}",
                "file": filename,
            }), 500
        elif type(result) is str:
            return jsonify({
                "status": "200",
                "message": "success",
                "file": result,
            }), 200
    except Exception as err:
        logging.info(str(err))
        return jsonify({
            "status": "500",
            "message": f"encountered error: {str(err)}",
            "file": filename,
        }), 500

if __name__ == '__main__':
    app.run(
        debug=True,
        host='0.0.0.0',
        port=8000,
    )