defmodule Scrapers.ProjectEuler do
  @moduledoc """
  This module fetches pages from Project Euler.
  """
  use HTTPoison.Base

  @base_url "https://projecteuler.net/problem="
  # @number_of_problems 624
  @number_of_problems 3
  # Milliseconds to wait between fetches
  @delay 1000 * 2

  def fetch_puzzle(id) do
    url = create_url(id)

    case HTTPoison.get(url) do
      {:ok, %HTTPoison.Response{status_code: 200, body: body}} -> body
      {:ok, %HTTPoison.Response{status_code: 404}} -> "Not found :("
      {:error, %HTTPoison.Error{reason: reason}} -> reason
    end
  end

  def process_response_body(body) do
    {:ok, document} = Floki.parse_document(body)

    [{"h2", _, [problem_name]} | _] = Floki.find(document, "#content h2")
    [{"div", _, problem_content} | _] = Floki.find(document, "div.problem_content")
    [{"h3", _, [problem_info]} | _] = Floki.find(document, "div#problem_info h3")
    regex = ~r"Problem (\d+)"
    [_, id] = Regex.run(regex, problem_info)

    %{
      id: id,
      title: problem_name,
      url: create_url(id),
      body: Floki.raw_html(problem_content)
    }
  end

  # TODO: write this function
  def save_json_file(j) do
    IO.inspect(j)
    :timer.sleep(@delay)
    j
  end

  def get_all_problems() do
    Enum.map(1..@number_of_problems, &fetch_puzzle/1)
    |> Enum.map(&process_response_body/1)
    |> Enum.map(&save_json_file/1)
  end

  defp create_url(id), do: "#{@base_url}#{id}"
end
