---
title: Creating My Site
date: 8/27/2021
image: site.webp
draft: false
---

I bought my first domain through [Google Domains](https://domains.google.com) and actually hosted my site through Google's website builder. It really is a great free builder with solid <abbr title="Search Engine Optimization">SEO</abbr>, but I'm a programmer, so I wanted to program.

Distributed networks are an amazing technology, and I was drawn to IPFS hosting. I used [Pinata](https://www.pinata.cloud/) to pin my new site, but when I tried using an HTTP request for a different page, it _failed_. Instead of fetching pages dynamically, I created a site that downloaded every page on load. This was **not** scalable. The method I ended up using for posts was incredibly simple, though. All I needed to do was create a javascript file and import it into `main.js`.

While still hosting on IPFS, I connected my GitHub <abbr title="Repository">repo</abbr> to Pinata and enjoyed automatic deployment; however, this was nearing my interest in [Go](https://go.dev). I saw how effective Go was for serving, so I decided to take my site off the IPFS network and deploy with Go.

Now that I had a house-hold server hosting my website, I could begin to move toward a scalable design. Slowly but surely, I converted all posts to pure <abbr title="Markdown">.md</abbr> files and handled `GET` requests from the server. At that point I had a fully functioning Go server, but I wasn't finished yet.

I moved from [splch.dev](https://splch.dev) to [splch.net](https://splch.net) and finally settled with [slc.is](https://slc.is). After learning more about <abbr title="Top Layer Domain">TLD</abbr>s, I realized I wouldn't want to have a country with lax laws ultimately control my website. After searching for countries with the [strongest privacy laws](/#The%20Best%20TLD%20is%20Not%20.com), I settled with Iceland's `.is` TLD. With this strong domain name, I ventured forward in refining the website.

Having to create new markdown files on the server was tedious and the website was created to simplify content creation. With a server, I could finally create posts from the website! I made it so I could click on the <abbr title="SPencer Lee CHurchill">SPLCH</abbr> logo and a markdown editor would appear _(I'm using that very editor to write this post)_. With a simple `POST` request, I verify and write the submitted post to a markdown file. In addition, the server returns markdown posts from a folder, so all that was required was to save the new markdown file in that `posts/` folder and it would populate in the website.

I'm sure more will come with the site, but having a built-in editor makes creating posts like these a breeze. To make the experience nicer, I started using [EasyMDE](https://easy-markdown-editor.tk/) to create and edit posts, and [Marked](https://marked.js.org/) to render the markdown to <abbr title="HyperText Markup Language">HTML</abbr>. Then, [DOMPurify](https://cure53.de/purify) sanitizes the HTML before populating. I also implemented [KaTex](https://katex.org/) for math and [highlight.js](https://highlightjs.org/) for code appearances:

Edit (9/6/2021) - I removed EasyMDE and highlight.js to decrease load times. I now render code with a replit iFrame.

<iframe class="container" width="100%" frameborder="0" src="https://replit.com/@splch/hadamard?lite=1"></iframe>

```python
def main():
	msg = "Hadamard Gate:"
	print(msg)

main()
```

```
$$
\frac{1}{\sqrt{2}}
\begin{bmatrix}
 1 & 1 \\
 1 & -1
\end{bmatrix}
$$
```

It's fun to have this framework for posting — I'm having a more positive experience using <abbr title="Content Delivery Network">CDN</abbr>s to host the external libraries. Not having to worry about updating libraries I'm using is relaxing as I can focus on improving my own scripts and posts. Clearly, I'm excited about this website and hope to share it with more people! I want to make this code open-source and create a helpful template site; however, I'll first need to improve the code a bit. But I'm very happy to say that the major work is done and the functionality is here! 🥳
