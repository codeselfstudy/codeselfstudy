defmodule Scrapers.ProjectEuler do
  @moduledoc """
  This module fetches pages from Project Euler.
  """
  use Hound.Helpers

  @base_url "https://projecteuler.net/problem="
  @number_of_problems 624

  def run do
    Hound.start_session()

    navigate_to("#{@base_url}1")
    IO.inspect(page_title())

    # Automatically invoked if the session owner process crashes
    Hound.end_session()
  end

  # def get_all_problems() do
  #   Enum.map(1..@number_of_problems, &fetch_problem/1)
  # end
end
