defmodule ScrapersTest do
  use ExUnit.Case
  doctest Scrapers

  test "greets the world" do
    assert Scrapers.hello() == :world
  end
end
