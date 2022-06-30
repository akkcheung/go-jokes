This is a clone of [Remix Jokes](https://remix-jokes.lol/jokes) using golang and [htmx](https://htmx.org) only.

General features:

- Sign In 
- Sign Up for new user
- add your own joke(s)
- get a random joke

Installation (deploy local):

1. docker build -t go-jokes .
2. docker run --rm -p 5000:5000 go-jokes

Installation (deploy to Heroku):

1. heroku login 
2. heroku create
3. heroku stack:set container
4. git push heroku main


Application design:

- MVC design
- 0% Javascript
- Just Golang, no web framework is used
- Render html using Goland HTML Template

Reference(s):

- [The Hypermedia Drive Application(HDA) architecture](https://htmx.org/essays/hypermedia-driven-applications/)
- use Hypermedia as the Engine of Applcation State [hateoas](https://en.wikipedia.org/wiki/HATEOAS)
- [Deploy goland webapp on heroku](https://dzone.com/articles/deploying-a-simple-golang-webapp-on-heroku)
