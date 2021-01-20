import 'buefy/dist/buefy.css'

import './etabs.scss'

declare global {
  interface Window {
    openExternal?: (url: string) => void;
  }
}

const navEl = document.querySelector('nav') as HTMLElement
const tabEl = document.querySelector('nav > ul') as HTMLUListElement
const mainEl = document.querySelector('main') as HTMLElement
const originalOpen = window.open

navEl.classList.remove('tabs')
tabEl.style.display = 'none'

window.open = function (url = '', title = '') {
  if (!url.startsWith('/#/')) {
    if (window.openExternal) {
      window.openExternal(url)
      return null
    }

    return originalOpen(url, '_blank', 'noopener noreferrer')
  }

  const li = document.createElement('li')
  li.className = 'is-active'

  const liA = document.createElement('a')
  li.append(liA)
  liA.innerText = title
  liA.setAttribute('role', 'button')
  liA.onclick = () => {
    const index = Array.from(tabEl.querySelectorAll('li > a')).indexOf(liA)

    tabEl.querySelectorAll('li').forEach((el, i) => {
      if (i !== index) {
        el.classList.remove('is-active')
      } else {
        el.classList.add('is-active')
      }
    })

    mainEl.querySelectorAll('iframe').forEach((el, i) => {
      if (i !== index) {
        el.style.display = 'none'
      } else {
        el.style.display = 'block'
      }
    })
  }

  if (tabEl.querySelector('li')) {
    const liAClose = document.createElement('a')
    liAClose.className = 'delete is-small'
    liAClose.onclick = () => {
      let i = Array.from(tabEl.querySelectorAll('li > a')).indexOf(liA)
      if (i < 1) {
        return
      }

      const li = Array.from(tabEl.querySelectorAll('li'))[i]
      if (li.classList.contains('is-active')) {
        setTimeout(() => {
          i--

          Array.from(tabEl.querySelectorAll('li'))[i].classList.add('is-active')
          Array.from(mainEl.querySelectorAll('iframe'))[i].style.display = ''
        }, 10)
      }

      li.remove()

      const iframe = Array.from(mainEl.querySelectorAll('iframe'))[i]
      iframe.remove()

      if (Array.from(tabEl.querySelectorAll('li')).length <= 1) {
        navEl.classList.remove('tabs')
        tabEl.style.display = 'none'
      }
    }

    liA.append(liAClose)
  }

  tabEl.querySelectorAll('li').forEach((el) => {
    el.classList.remove('is-active')
  })

  tabEl.append(li)

  if (Array.from(tabEl.querySelectorAll('li')).length > 1) {
    navEl.classList.add('tabs')
    tabEl.style.display = ''
  }

  const iframe = document.createElement('iframe')
  iframe.src = url

  mainEl.querySelectorAll('iframe').forEach((el) => {
    el.style.display = 'none'
  })

  mainEl.append(iframe)

  return iframe.contentWindow
}

open('/#/', 'Home')
