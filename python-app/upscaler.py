import logging
import os.path

import torch
import numpy as np
from typing import Optional, Union
from PIL import Image
from basicsr.archs.rrdbnet_arch import RRDBNet
from realesrgan import RealESRGANer

class Upscaler:
    def __init__(self, model_path: str, scale: int):
        self.model_path: str = model_path
        self.scale: int = scale
        self.upsampler: Optional[RealESRGANer] = None

    def new_default(self) -> None:
        """Creates a new default upsampler model and assigns it to the Upscaler"""
        if self.upsampler is None:
            state_dict = torch.load(self.model_path, map_location=torch.device('cpu'))['params_ema']
            model = RRDBNet(num_in_ch=3, num_out_ch=3, num_feat=64, num_block=23, num_grow_ch=32, scale=self.scale)
            model.load_state_dict(state_dict, strict=True)

            self.upsampler = RealESRGANer(
                scale=self.scale,
                model_path=self.model_path,
                model=model,
                tile=0,
                pre_pad=0,
                half=False,
            )
        else:
            logging.info("upsampler is already assigned!")

    def process(self, filename: str) -> Union[str, Exception]:
        """Does the upscaling business and saves to ./uploads"""
        name, ext = os.path.splitext(filename)
        if self.upsampler is None:
            err = RuntimeError("upsampler is not initialized! consider calling new_default() ?")
            logging.error(err)
            return err
        try:
            img = np.array(Image.open(filename).convert('RGB'))
            output, _ = self.upsampler.enhance(img, outscale=4)

            output_img = Image.fromarray(output)
            new_filename = f'{name.split("/")[-1]}_upscaled{ext}'
            output_img.save(f'./uploads/{new_filename}')
            return new_filename
        except Exception as err:
            logging.error(str(err))
            return err