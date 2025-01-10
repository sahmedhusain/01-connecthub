# Forum

This project is a web-based forum application that allows users to create, view, and manage posts. It includes features such as user authentication, image uploads, moderation tools, and more.

## Features

- User authentication and authorization
- Post creation and management
- Image uploads
- Moderation tools
- Security features
- Advanced forum functionalities

## Getting Started

### Prerequisites

- Docker
- Go (Golang)
- MySQL

### Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/forum.git
    cd forum
    ```

2. Build and run the Docker containers:
    ```sh
    docker-compose up --build
    ```

3. Access the application at `http://localhost:8080`.

## Example Usage

1. Register a new user.
2. Log in with the registered user.
3. Create a new post.
4. View and interact with posts.

## Limitations

- Limited to basic forum functionalities.
- No real-time updates.

## Future Improvements

- Add real-time notifications.
- Improve the user interface.
- Implement more advanced moderation tools.

## Contributing

1. Fork the repository.
2. Create your feature branch (`git checkout -b feature/NewFeature`).
3. Commit your changes (`git commit -m 'Add NewFeature'`).
4. Push to the branch (`git push origin feature/NewFeature`).
5. Open a Pull Request.

## Authors

- Sayed Ahmed Husain
- Qasim Aljaffer
- Mohammed AlAlawi
- Abdulla Alasmawi

## License

This project is licensed under the MIT License. See the [LICENSE.md](http://_vscodecontentref_/1) file for details.

## Project Structure

Forum/
├── .dockerignore
├── .DS_Store
├── .gitignore
├── compose.yaml
├── database/
│   ├── database.go
│   └── Sqlstat.go
├── dockerfile
├── erd/
│   ├── Forum.mwb
│   └── ForumDB.sql
├── go.mod
├── go.sum
├── LICENSE.md
├── main.go
├── README.md
├── run.sh
├── src/
│   ├── .DS_Store
│   ├── advanced-features/
│   ├── authentication/
│   ├── image-upload/
│   ├── moderation/
│   ├── security/
│   └── server/
├── static/
│   ├── .DS_Store
│   ├── assets/
│   ├── css/
│   ├── js/
│   └── uploads/
└── templates/
    ├── admin.html
    ├── changepassword.html
    ├── error.html
    ├── home.html
    ├── index.html
    ├── login.html
    ├── moderator.html
    ├── myprofile.html
    ├── newpost.html
    ├── notifications.html
    └── post.html