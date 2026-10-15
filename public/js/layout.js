'use strict';

// Regra #2: todo SVG leva width/height explícitos (sem eles renderiza em 300x150).
// Navegação: 20x20. Ações (editar, excluir, confirmar, incluir): 16x16.
const ICONES_SVG = {
  dashboard: {
    tamanho: 20,
    corpo: '<rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/>',
  },
  tag: {
    tamanho: 20,
    corpo: '<path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/>',
  },
  'credit-card': {
    tamanho: 20,
    corpo: '<rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/>',
  },
  users: {
    tamanho: 20,
    corpo: '<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>',
  },
  dollar: {
    tamanho: 20,
    corpo: '<line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>',
  },
  menu: {
    tamanho: 20,
    corpo: '<line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="18" x2="21" y2="18"/>',
  },
  x: {
    tamanho: 20,
    corpo: '<line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>',
  },
  logout: {
    tamanho: 20,
    corpo: '<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/>',
  },
  edit: {
    tamanho: 16,
    corpo: '<path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"/>',
  },
  trash: {
    tamanho: 16,
    corpo: '<polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6"/><path d="M14 11v6"/><path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/>',
  },
  check: {
    tamanho: 16,
    corpo: '<polyline points="20 6 9 17 4 12"/>',
  },
  plus: {
    tamanho: 16,
    corpo: '<line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>',
  },
};

// Retorna o SVG inline do ícone; sem o 2º argumento usa o tamanho padrão do ícone.
function svgIcon(nome, tamanho) {
  const icone = ICONES_SVG[nome];
  if (!icone) return '';

  const t = Number(tamanho) || icone.tamanho;
  return `<svg xmlns="http://www.w3.org/2000/svg" width="${t}" height="${t}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">${icone.corpo}</svg>`;
}

const LINKS_MENU = [
  { href: 'index.html', rotulo: 'Dashboard', icone: 'dashboard' },
  { href: 'categoria.html', rotulo: 'Categorias', icone: 'tag' },
  { href: 'despesa.html', rotulo: 'Despesas', icone: 'credit-card' },
  { href: 'usuario.html', rotulo: 'Usuários', icone: 'users', somenteAdmin: true },
];

const CLASSE_LINK = 'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium';
const CLASSE_LINK_ATIVO = 'bg-indigo-50 text-indigo-700';
const CLASSE_LINK_INATIVO = 'text-slate-600 hover:bg-slate-50 hover:text-slate-900';

const CLASSE_NIVEL = {
  ADMIN: 'bg-indigo-100 text-indigo-700',
  READ: 'bg-gray-100 text-gray-600',
};

// Injeta sidebar e topbar mobile em #app, movendo o conteúdo original da página para dentro do layout.
function montarLayout() {
  const app = document.getElementById('app');
  if (!app || !Auth.getToken()) return;

  const usuario = Auth.getUser();
  const pagina = location.pathname.split('/').pop() || 'index.html';
  const links = LINKS_MENU
    .filter((l) => !l.somenteAdmin || Auth.isCompleto())
    .map((l) => {
      const classe = `${CLASSE_LINK} ${l.href === pagina ? CLASSE_LINK_ATIVO : CLASSE_LINK_INATIVO}`;
      const atual = l.href === pagina ? ' aria-current="page"' : '';
      return `<a href="${window.BASE_PATH}/${l.href}" class="${classe}"${atual}>${svgIcon(l.icone)}<span>${l.rotulo}</span></a>`;
    })
    .join('');

  const conteudoOriginal = Array.from(app.childNodes);

  app.className = 'min-h-screen bg-slate-50';
  app.innerHTML = `
    <div id="sidebar-overlay" class="fixed inset-0 z-30 hidden bg-slate-900/40 lg:hidden"></div>

    <aside id="sidebar" class="fixed inset-y-0 left-0 z-40 flex w-64 -translate-x-full flex-col border-r border-gray-100 bg-white transition-transform duration-200 lg:translate-x-0">
      <div class="flex h-16 items-center gap-2 border-b border-gray-100 px-5">
        <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-600 text-white">${svgIcon('dollar')}</span>
        <span class="flex-1 text-base font-semibold text-slate-800">Registro de Despesas</span>
        <button id="btn-fechar-menu" type="button" aria-label="Fechar menu" class="rounded-lg p-1 text-slate-500 hover:bg-slate-100 lg:hidden">${svgIcon('x')}</button>
      </div>

      <nav class="flex-1 space-y-1 overflow-y-auto px-3 py-4">${links}</nav>

      <div class="border-t border-gray-100 p-4">
        <div class="flex items-center gap-3">
          <div id="usuario-avatar" class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-indigo-100 text-sm font-semibold text-indigo-700"></div>
          <div class="min-w-0 flex-1">
            <p id="usuario-nome" class="truncate text-sm font-medium text-slate-800"></p>
            <span id="usuario-nivel" class="mt-0.5 inline-block rounded-full px-2 py-0.5 text-xs font-medium"></span>
          </div>
        </div>
        <button id="btn-sair" type="button" class="mt-3 flex w-full items-center justify-center gap-2 rounded-lg border border-gray-200 px-3 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50">
          ${svgIcon('logout')}<span>Sair</span>
        </button>
      </div>
    </aside>

    <div class="lg:pl-64">
      <header class="sticky top-0 z-20 flex h-14 items-center gap-3 border-b border-gray-100 bg-white px-4 lg:hidden">
        <button id="btn-menu" type="button" aria-label="Abrir menu" class="rounded-lg p-1.5 text-slate-600 hover:bg-slate-100">${svgIcon('menu')}</button>
        <span class="font-semibold text-slate-800">Registro de Despesas</span>
      </header>
      <main id="conteudo-pagina" class="p-4 sm:p-6 lg:p-8"></main>
    </div>`;

  // Dados do usuário entram via textContent: nome e e-mail nunca passam por innerHTML.
  const nome = usuario.nome || usuario.email;
  document.getElementById('usuario-avatar').textContent = (nome.trim().charAt(0) || '?').toUpperCase();
  const elNome = document.getElementById('usuario-nome');
  elNome.textContent = nome;
  elNome.title = usuario.email;
  const elNivel = document.getElementById('usuario-nivel');
  elNivel.textContent = usuario.nivel;
  elNivel.className += ` ${CLASSE_NIVEL[usuario.nivel] || CLASSE_NIVEL.READ}`;

  const conteudo = document.getElementById('conteudo-pagina');
  conteudoOriginal.forEach((no) => conteudo.appendChild(no));

  configurarMenuMobile();
  document.getElementById('btn-sair').addEventListener('click', () => Auth.logout());
}

function configurarMenuMobile() {
  const sidebar = document.getElementById('sidebar');
  const overlay = document.getElementById('sidebar-overlay');

  const alternar = (aberto) => {
    sidebar.classList.toggle('-translate-x-full', !aberto);
    overlay.classList.toggle('hidden', !aberto);
    document.body.classList.toggle('overflow-hidden', aberto);
  };

  document.getElementById('btn-menu').addEventListener('click', () => alternar(true));
  document.getElementById('btn-fechar-menu').addEventListener('click', () => alternar(false));
  overlay.addEventListener('click', () => alternar(false));
  sidebar.querySelectorAll('nav a').forEach((a) => a.addEventListener('click', () => alternar(false)));
  document.addEventListener('keydown', (e) => e.key === 'Escape' && alternar(false));
  window.matchMedia('(min-width: 1024px)').addEventListener('change', (e) => e.matches && alternar(false));
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', montarLayout);
} else {
  montarLayout();
}
