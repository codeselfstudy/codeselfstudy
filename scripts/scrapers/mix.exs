defmodule Scrapers.MixProject do
  use Mix.Project

  def project do
    [
      app: :scrapers,
      version: "0.1.0",
      elixir: "~> 1.9",
      start_permanent: Mix.env() == :prod,
      deps: deps()
    ]
  end

  # Run "mix help compile.app" to learn about applications.
  def application do
    [
      extra_applications: [:logger],
      mod: {Scrapers.Application, []}
    ]
  end

  # Run "mix help deps" to learn about dependencies.
  defp deps do
    [
      {:hound, "~> 1.0"},
      {:floki, "~> 0.24.0"},
      {:httpoison, "~> 1.6"}
    ]
  end
end
