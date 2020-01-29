defmodule Scrapers.ProjectEuler do
  @moduledoc """
  This module fetches pages from Project Euler.
  """
  use HTTPoison.Base

  @base_url "https://projecteuler.net/problem="
  # TODO: this may have changed:
  @output_dir File.cwd!() <> "/../containers/mongo/seed_data/project_euler"
  # Milliseconds to wait between fetches
  @delay 1000 * 1

  def fetch_puzzle(id) do
    url = create_url(id)
    IO.puts("fetching #{url}")

    :timer.sleep(@delay)

    case HTTPoison.get(url) do
      {:ok, %HTTPoison.Response{status_code: 200, body: body}} -> body
      {:ok, %HTTPoison.Response{status_code: 404}} -> "Not found :("
      {:error, %HTTPoison.Error{reason: reason}} -> reason
      x -> IO.inspect(x)
    end
  end

  def process_response_body(body) do
    {:ok, document} = Floki.parse_document(body)

    [{"h2", _, [problem_name | _]} | _] = Floki.find(document, "#content h2")
    [{"div", _, problem_content} | _] = Floki.find(document, "div.problem_content")
    [{"h3", _, [problem_info | _]} | _] = Floki.find(document, "div#problem_info h3")
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
  def save_json_file(%{:id => id} = m) do
    IO.inspect(m)
    filename = @output_dir <> "/" <> id <> ".json"
    content = Jason.encode!(m)
    File.write(filename, content)
    IO.puts("wrote #{filename}")
  end

  def download_problems(start_id, end_id) do
    Enum.map(@start_id..@end_id, &fetch_puzzle/1)
    |> Enum.map(&process_response_body/1)
    |> Enum.map(&save_json_file/1)
  end

  defp create_url(id), do: "#{@base_url}#{id}"
end
