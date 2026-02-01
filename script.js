// script.js
// Progressive enhancement for site interactions. Does not modify markup structure.

(function(){
  'use strict';
  window.__scriptJsLoaded = true;

  const $ = (sel, all=false) => all ? Array.from(document.querySelectorAll(sel)) : document.querySelector(sel);

  // Smooth scroll for anchor links
  document.addEventListener('click', (e) => {
    const a = e.target.closest('a');
    if(!a) return;
    const href = a.getAttribute('href');
    if(href && href.startsWith('#')){
      const id = href.slice(1);
      const target = document.getElementById(id);
      if(target){
        e.preventDefault();
        target.scrollIntoView({behavior:'smooth', block:'start'});
        // Preserve back button functionality
        if(history.pushState) history.pushState(null, '', '#'+id);
      }
    }
  });

  // Mobile menu
  const menuToggle = $('.menu-toggle');
  const mobileMenu = $('#mobile-menu');
  if(menuToggle && mobileMenu){
    menuToggle.addEventListener('click', ()=>{
      const expanded = menuToggle.getAttribute('aria-expanded') === 'true';
      menuToggle.setAttribute('aria-expanded', String(!expanded));
      if(mobileMenu.hasAttribute('hidden')){
        mobileMenu.removeAttribute('hidden');
        menuToggle.textContent = 'Close';
      } else{
        mobileMenu.setAttribute('hidden','');
        menuToggle.textContent = 'Menu';
      }
    });
  }

  // Auto-close menu after navigation
  if(mobileMenu){
    mobileMenu.addEventListener('click', (e)=>{
      const a = e.target.closest('a');
      if(a){
        mobileMenu.setAttribute('hidden','');
        menuToggle.setAttribute('aria-expanded','false');
        menuToggle.textContent = 'Menu';
      }
    });
  }

  // Theme toggle with localStorage persistence
  const themeToggle = $('#theme-toggle');
  const root = document.documentElement;
  const savedTheme = localStorage.getItem('naevermore-theme');

  if(savedTheme) root.setAttribute('data-theme', savedTheme);

  if(themeToggle){
    const isDark = (root.getAttribute('data-theme') === 'dark');
    themeToggle.textContent = isDark ? 'Light Mode' : 'Dark Mode';
    themeToggle.setAttribute('aria-pressed', String(isDark));
    themeToggle.addEventListener('click', ()=>{
      const current = root.getAttribute('data-theme');
      const next = (current === 'dark') ? '' : 'dark';
      if(next) root.setAttribute('data-theme', next); else root.removeAttribute('data-theme');
      localStorage.setItem('naevermore-theme', next);
      const isNowDark = (next === 'dark');
      themeToggle.textContent = isNowDark ? 'Light Mode' : 'Dark Mode';
      themeToggle.setAttribute('aria-pressed', String(isNowDark));
    });
  }

  // Gallery lightbox
  const thumbs = $('.thumb', true);
  const lightbox = $('#lightbox');
  const lightboxImg = $('#lightbox-img');
  const lbClose = $('#lightbox-close');

  function openLightbox(src, alt=''){
    if(!lightbox) return;
    lightboxImg.src = src;
    lightboxImg.alt = alt;
    lightbox.removeAttribute('hidden');
    lightbox.setAttribute('aria-hidden','false');
  }
  function closeLightbox(){
    if(!lightbox) return;
    lightbox.setAttribute('hidden','');
    lightbox.setAttribute('aria-hidden','true');
    lightboxImg.src = '';
  }

  thumbs.forEach(btn=>{
    btn.addEventListener('click', ()=>{
      const src = btn.dataset.full;
      const alt = btn.querySelector('img')?.alt || '';
      openLightbox(src, alt);
    });
  });

  if(lbClose) lbClose.addEventListener('click', closeLightbox);
  if(lightbox) lightbox.addEventListener('click', (e)=>{ if(e.target === lightbox) closeLightbox(); });
  document.addEventListener('keydown', (e)=>{ if(e.key === 'Escape') closeLightbox(); });

  // Contact form (demo only - no backend submission)
  const form = $('#contact-form');
  const formMsg = $('#form-msg');
  if(form){
    form.addEventListener('submit', (e)=>{
      e.preventDefault();
      const data = new FormData(form);
      const name = data.get('name')?.toString().trim();
      const email = data.get('email')?.toString().trim();
      const message = data.get('message')?.toString().trim();

      if(!name || !email || !message){
        formMsg.textContent = 'Please complete all fields.';
        formMsg.style.color = 'var(--accent)';
        return;
      }

      formMsg.textContent = 'Thanks — your message has been queued (demo only).';
      formMsg.style.color = 'var(--muted)';
      form.reset();
    });
  }

  // Booking form submits directly to Formspree (no JS intervention needed)

  // Auto-update copyright year
  const year = $('#year'); if(year) year.textContent = new Date().getFullYear();

  // Hide scroll hint after user scrolls
  const scrollArrow = document.querySelector('.scroll-down-arrow');
  if(scrollArrow){
    function toggleArrow(){
      const y = window.scrollY || document.documentElement.scrollTop || 0;
      if(y > 20){ scrollArrow.classList.add('is-hidden'); }
      else{ scrollArrow.classList.remove('is-hidden'); }
    }
    toggleArrow();
    window.addEventListener('scroll', toggleArrow, { passive: true });
    window.addEventListener('resize', toggleArrow);
  }

  // Site search with cross-page navigation
  const search = $('#site-search');
  if(search){
    // Static content map enables searching pages before they're loaded
    const siteContent = [
      { page: 'Nævermore.html', section: 'Home', keywords: 'home landing main page nessie audio' },
      { page: 'Nævermore.html', section: 'Welcome to Nessie Audio', keywords: 'welcome about professional audio production services music recording mixing mastering' },
      { page: 'Nævermore.html', section: 'News', keywords: 'news updates announcements latest' },
      { page: 'Nævermore.html', section: 'Follow Us', keywords: 'follow social media instagram facebook youtube tiktok' },
      { page: 'music.html', section: 'Portfolio', keywords: 'portfolio music audio songs tracks production work catalogue bluegrass classical cinematic contemporary country edm indie metal punk rap rock' },
      { page: 'merch.html', section: 'Merch', keywords: 'merch merchandise store shop buy products' },
      { page: 'nessie-digital.html', section: 'Nessie Digital', keywords: 'nessie digital services' }
    ];

    let dropdown = document.querySelector('.search-dropdown');
    if(!dropdown){
      dropdown = document.createElement('div');
      dropdown.className = 'search-dropdown';
      dropdown.setAttribute('role', 'listbox');
      search.parentElement.style.position = 'relative';
      search.parentElement.appendChild(dropdown);
    }

    search.addEventListener('input', ()=>{
      const q = search.value.trim().toLowerCase();

      if(q === ''){
        dropdown.innerHTML = '';
        dropdown.style.display = 'none';
        return;
      }

      // Search visible content on current page
      const currentPageResults = [];
      const searchableElements = document.querySelectorAll('main h1, main h2, main h3, main h4, main p, main li:not(.main-nav li):not(.mobile-menu li), main a:not(.main-nav a):not(.mobile-menu a), main label');

      searchableElements.forEach(el=>{
        const text = el.innerText || el.textContent;
        if(text && text.toLowerCase().includes(q)){
          let section = el.closest('section, aside, .show-promo-row');
          let sectionTitle = null;
          if(section){
            const heading = section.querySelector('h1, h2, h3');
            if(heading) sectionTitle = heading.innerText.trim();
          }

          if(sectionTitle){
            currentPageResults.push({
              section: sectionTitle,
              element: el,
              isCurrentPage: true
            });
          }
        }
      });

      // Search static content map for other pages
      const siteResults = siteContent.filter(item=>{
        return item.section.toLowerCase().includes(q) || item.keywords.toLowerCase().includes(q);
      });

      // Deduplicate results
      const allResults = [];
      const seen = new Set();

      currentPageResults.forEach(r=>{
        const key = r.section;
        if(!seen.has(key)){
          seen.add(key);
          allResults.push(r);
        }
      });

      siteResults.forEach(r=>{
        const key = r.section + r.page;
        if(!seen.has(key)){
          seen.add(key);
          allResults.push(r);
        }
      });

      if(allResults.length > 0){
        dropdown.innerHTML = allResults.slice(0, 8).map((result, i)=>{
          return `<div class="search-result-item" data-index="${i}">
            <div class="result-section">${result.section}</div>
          </div>`;
        }).join('');
        dropdown.style.display = 'block';

        dropdown.querySelectorAll('.search-result-item').forEach((item, i)=>{
          item.addEventListener('click', ()=>{
            const result = allResults[i];
            dropdown.style.display = 'none';
            search.value = '';

            if(result.isCurrentPage){
              result.element.scrollIntoView({ behavior: 'smooth', block: 'center' });
              // Brief highlight to show location
              result.element.style.backgroundColor = 'rgba(192, 192, 192, 0.3)';
              setTimeout(()=> result.element.style.backgroundColor = '', 2000);
            } else {
              window.location.href = result.page;
            }
          });
        });
      } else {
        dropdown.innerHTML = '<div class="search-no-results">No results found</div>';
        dropdown.style.display = 'block';
      }
    });

    document.addEventListener('click', (e)=>{
      if(!search.contains(e.target) && !dropdown.contains(e.target)){
        dropdown.style.display = 'none';
      }
    });

    search.addEventListener('keydown', (e)=>{
      if(e.key === 'Enter'){
        e.preventDefault();
        const firstResult = dropdown.querySelector('.search-result-item');
        if(firstResult) firstResult.click();
      }
      if(e.key === 'Escape'){
        dropdown.style.display = 'none';
        search.value = '';
      }
    });
  }

  // Video loader - populates <video> from data-video attribute, with fallback for unsupported browsers
  Array.from(document.querySelectorAll('.video-wrap')).forEach(el=>{
    const file = (el.getAttribute('data-video') || '').trim();
    const poster = (el.getAttribute('data-poster') || '').trim();
    const section = el.closest('.video-section');
    if(!file){ if(section) section.style.display = 'none'; return; }

    let video = el.querySelector('video');
    if(!video){
      video = document.createElement('video');
      video.className = 'site-video';
      video.setAttribute('controls', '');
      video.setAttribute('playsinline', '');
      video.preload = 'metadata';
      el.insertBefore(video, el.firstChild);
    }

    let source = video.querySelector('source');
    if(source){ source.src = file; }
    else{
      source = document.createElement('source');
      source.src = file;
      // MIME type helps browsers decide if they can play before downloading
      const ext = file.split('.').pop().toLowerCase();
      if(ext === 'mp4') source.type = 'video/mp4';
      if(ext === 'webm') source.type = 'video/webm';
      video.appendChild(source);
    }

    if(poster) video.poster = poster;
    const fallback = el.querySelector('.video-fallback');
    const fallbackLink = el.querySelector('#video-fallback-link');
    if(fallbackLink){
      fallbackLink.href = file;
      fallbackLink.textContent = 'Open video file in new tab';
      fallbackLink.setAttribute('aria-label', 'Open video file');
    }

    function showFallback(){
      if(!fallback) return;
      if(!el.querySelector('.video-thumb') && video.poster){
        const a = document.createElement('a');
        a.className = 'video-thumb';
        a.href = file;
        a.target = '_blank';
        a.rel = 'noopener noreferrer';
        const img = document.createElement('img');
        img.src = video.poster;
        img.alt = 'Open video in new tab';
        a.appendChild(img);
        fallback.insertBefore(a, fallback.firstChild);
      }
      fallback.hidden = false;
      video.hidden = true;
    }

    video.addEventListener('canplay', ()=>{ if(fallback) fallback.hidden = true; video.hidden = false; });
    video.addEventListener('error', ()=>{ showFallback(); });

    try{ video.load(); }
    catch(e){ showFallback(); }
  });

})();
