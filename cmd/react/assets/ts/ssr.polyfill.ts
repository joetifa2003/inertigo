import "fastestsmallesttextencoderdecoder/EncoderDecoderTogether.min.js";
import { FormData } from "formdata-polyfill/esm.min.js";
import { URL, URLSearchParams } from "whatwg-url";

globalThis.FormData = FormData;
globalThis.URL = URL;
globalThis.URLSearchParams = URLSearchParams;
