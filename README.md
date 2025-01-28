I set out to build a project that would use Python and Go together, and this was the result.

It's an image upscaler thats entirely docker-based, and contains a GO API with a Postgres DB that talks with to a Python image upscaling script to upscale images.

----------------------------------------------------------------
In order for the upscaler to work, a pre-trained model is needed.

This app was built using RealESRGAN_x4plus.pth, which you can get here:
https://github.com/xinntao/Real-ESRGAN

CTRL+F for "RealESRGAN_x4plus.pth" on the page and you'll see a download link.

Once you have the model, just place the .pth file in the `./python-app/models` directory right next to `readme.md`.

Note that using a different model will require tweaks in the application's code.
----------------------------------------------------------------
You can run `make build` within the terminal inside the app dir, and it should build the docker containers for you.
`make up` starts the application, `make down` shuts it down.

You can send a GET request to "/" and you'll receive the available commands.
Common usage goes as follows:
1. Use POST to `/upload` to upload a file
2. Use GET to `/files` to see files and their IDs
3. Use POST to `/upscale/<id>` to upscale the image by a factor of 4 (by default)
4. Profit ???
