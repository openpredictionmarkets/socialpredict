## Hypothetical Deploying from Dev

### Background

As of 31 May 2024, SocialPredict does not have a production version launched. That is to say, SocialPredict is designed for development, meaning out of the box, everything it comes with can be worked on within a local machine, but it's not set up to be served securely as a production instance on the web that can't be easily hacked.

That being said, while not recommended, it is possible to adapt the development version of SocialPredict to run in a production environment in its current state, for small and experimental deployments.

This document is geared at helping those who might want to convert the development version of socialpredict into a web deployment. While not all-encompassing and user caution is encouraged, this document can serve as a general guideline to get going.

### Contents

#### Setting Up Your Domain

First, you should purchase a domain, which you can do at any number of domain registrars. We are using Namecheap for this article.


### Set Up Digital Ocean Account

We are using Digital Ocean as a server, since it's predictable, easy and cheap. SocialPredict is designed to be performant, meaning it can perform well on a small server.

### Link Domain Directly to Digital Ocean

Once you have a domain name, reference your domain registrar's documentation for [connecting up your domain to a routing service such as aws Route53](https://www.namecheap.com/support/knowledgebase/article.aspx/10371/2208/how-do-i-link-my-domain-to-amazon-web-services/).


https://www.namecheap.com/support/knowledgebase/article.aspx/10375/2208/how-do-i-link-a-domain-to-my-digitalocean-account/

https://docs.digitalocean.com/products/networking/dns/getting-started/dns-registrars/