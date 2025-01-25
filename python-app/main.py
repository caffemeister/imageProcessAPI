from PIL.Image import Resampling
from PIL.ImageFile import ImageFile
from flask import Flask, request, jsonify, Response
from PIL import Image
import os

app = Flask(__name__)
UPSCALE_FACTOR: int = 3

@app.route('/upscale', methods=["POST"])
def upscale_file() -> tuple[Response, int]:
    # get filename from data
    data = request.get_json()
    filename = data.get('filename')

    print(type(filename))

    # check if filename exists
    if not filename:
        return jsonify({
            "status": 404,
            "message": "no file provided",
            "file": "",
        }), 404

    # upscale image
    try:
        file_path: str = f"./../uploads/{filename}"

        if not os.path.exists(file_path):
            return jsonify({
                "status": 404,
                "message": f"file {filename} not found",
                "file": filename,
            }), 404

        image: ImageFile = Image.open(file_path)
        width, height = image.size

        new_width: int = width*UPSCALE_FACTOR
        new_height: int = height*UPSCALE_FACTOR

        upscaled_image: Image = image.resize(size=(new_width, new_height), resample=Resampling.LANCZOS)
        upscaled_image_name: str = filename+"+up"
        upscaled_image.save(fp=f"./../uploads/{upscaled_image_name}")

        return jsonify({
            "status": "200",
            "message": "success",
            "file": upscaled_image_name,
        }), 200
    except Exception as err:
        return jsonify({
            "status": "500",
            "message": f"encountered error: {str(err)}",
            "file": filename,
        }), 500
