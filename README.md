
# MongoMove

A CLI tool built in GO that quickly moves records from a MongoDB source database to a target over two different servers

## Features

- Add source and target databases using config file
- Select source and target database
- Select one or more source and target collections
- Manage choices before starting copy
- Filter and pagination options on each table to aid selection

## Demo

- Select source and target database

  <img width="359" alt="image" src="https://github.com/user-attachments/assets/41a93e4d-a1b3-4c6a-b04b-ce93268be583">

- Select one or more source and target collections

<img width="466" alt="image" src="https://github.com/user-attachments/assets/88ff11af-dbc0-480d-afd9-e0cc729e4ba7">

- Manage choices before starting copy

<img width="409" alt="image" src="https://github.com/user-attachments/assets/ad225336-a02e-4ea0-ab21-a395a27702b7">

  
## Run Locally

Clone the project

```bash
  git clone https://github.com/CundyTech/mongo-move.git
```

Go to the project directory

```bash
  cd CundyTech/mongo-move
```

Install dependencies and build executable 

```bash
  go build
```

Run executable

```bash
  .\mongo-move.exe
```


## License

[MIT](https://choosealicense.com/licenses/mit/)

