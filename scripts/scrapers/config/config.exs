use Mix.Config

# I didn't figure out hound yet.
# https://github.com/HashNuke/hound/blob/master/notes/configuring-hound.md
config :hound, port: 5555, retries: 3, browser: "firefox"

import_config "#{Mix.env()}.exs"
