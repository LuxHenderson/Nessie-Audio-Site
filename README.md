# Nævermore — Static Site Scaffold

Quick scaffold for the Nævermore band site. Files created:

- `Nævermore.html` — Main HTML with sections, comments, and editable placeholders.
- `style.css` — Centralized variables, responsive layout, and dark mode styles.
- `script.js` — Smooth scroll, mobile menu, dark mode persistence, gallery lightbox, and simple contact demo.

How to use

1. Open `Nævermore.html` in your browser (double-click or use a Live Server extension).

2. Edit content quickly:
   - Band name / logo: in the header `.brand .logo` inside `Nævermore.html`.
   - Hero image: change the `background-image` in `.hero` in `style.css`.
   - Members: swap placeholder images and text in the `#about .members-grid`.
   - Streaming links: update anchors in `#music`.
   - Tour dates: edit the `#tour .events-list` items.
   - Gallery: replace thumbnail `src` and `data-full` attributes with your photos.

3. To change theme colors or type scale, update the CSS variables at the top of `style.css`.

Notes & next steps

- The contact form is a front-end demo and does not send emails. Add a server endpoint or a service (e.g. Formspree, Netlify Forms) to make it functional.
- For better performance, replace placeholder images with optimized assets and host them locally or via a CDN.
- If you want multiple pages, duplicate `Nævermore.html` into new HTML files and reuse `style.css` and `script.js`.

If you want, I can:
- Wire the contact form to a chosen backend or form service.
- Add a favicon and structured metadata (Open Graph, Twitter card).
- Improve accessibility and automated tests.

