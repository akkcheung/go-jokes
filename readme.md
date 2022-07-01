This is a clone of [Remix Jokes](https://remix-jokes.lol/jokes) using golang, [htmx](https://htmx.org) and [Spectre.css Framework](https://picturepan2.github.io/spectre/) only.

General features:

- Sign In 
- Sign Up for new user
- add your own joke(s)
- get a random joke

Application design:

- MVC design
- 0% Javascript means no build
- Just Golang, no web framework is used
- Render html using Golang html/template

Installation (local):

1. docker build -t go-jokes .
2. docker run --rm -p 5000:5000 go-jokes

Installation (deploy to Heroku):

1. heroku login 
2. heroku create
3. heroku stack:set container
4. git push heroku main
5. heroku open

Reference(s):

- [The Hypermedia Drive Application(HDA) architecture](https://htmx.org/essays/hypermedia-driven-applications/)
- [Deploy goland webapp on heroku](https://dzone.com/articles/deploying-a-simple-golang-webapp-on-heroku)
