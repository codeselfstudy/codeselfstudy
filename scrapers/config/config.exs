use Mix.Config

config :hound, port: 5555, retries: 3, browser: "firefox"

import_config "#{Mix.env()}.exs"
