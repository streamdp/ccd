const mq = window.matchMedia('(prefers-color-scheme: dark)')
if (mq.matches) {
    document.documentElement.classList.add('dark')
}
function toggleTheme() {
    const root = document.documentElement
    root.classList.toggle('dark')
    if (window.ccd) {
        const isDark = root.classList.contains('dark')
        window.ccd.changeTheme(isDark ? 'dark' : 'light')
    }
}