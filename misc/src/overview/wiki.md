# Wiki

The wiki is in a separate repo and auto-deploys on Netlify.

https://github.com/codeselfstudy/codeselfstudy_wiki

https://wiki.codeselfstudy.com/

These resources were useful in setting up the CI pipline:

- https://docs.travis-ci.com/user/deployment-v2/providers/netlify/
- https://docs.travis-ci.com/user/tutorial/
- https://github.com/travis-ci/travis.rb

Basically, Travis CI builds and deploys the site. A deploy key is encrypted and added with the Travis CI command-line interface (linked to above). The command `travis login` will connect the CLI to Travis, and then this command will add the encrypted Netlify token to the `.travis.yml` file.

```text
$ travis encrypt --add deploy.auth <auth>
```

(Netlify doesn't have to be connected to Github.)
