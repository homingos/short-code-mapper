# Short-Code-Mapper Utility

This is a functional utility to retrieve product short codes along with their names and calculated similarity score from l2 scores for a given question.

> results are stored both in   `JSON` and `CSV`

## Getting Started
- clone the repo
    ```bash
    git clone https://github.com/homingos/short-code-mapper
    ```

- install dependencies
    ```bash
    npm install
    ```

- add all <questions,answer> pairs in `questions.txt`.

- Set the environment(Env) as either `dev` or `prod`
In line 40; [milvus_dao_impl.go](./go-server/daos/milvus_dao_impl.go)

- Set the environment variables. Refer `go-server/.env.local` folder

- run the go server
    ```bash
    go run go-server/main.go
    ```


## Execute script

- Hit the below API endpoint first, in order to create a     `shortcode` to `name` mapping.
You need a valid `sitecode`
    ```html
    http://localhost:3000/generate-mappings/<sitecode>
    ```
> this would create a `mapping_<sitecode>.json`

- Next, run the client script for automated data creation
    ```sh
    node client
    ```

- check `short_code_output.csv` and `short_code_output.json` for results.
> Caution! The above two files will be appended for repeated runs. Duplicates are not handled. Erase content for any new run as needed.
