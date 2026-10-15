'use strict';

// Esconde os elementos .somente-admin para quem não é ADMIN. Só desfaz o que ela mesma escondeu,
// para não mostrar elementos que a página ocultou por outro motivo.
function aplicarPermissoes(raiz = document) {
  const admin = Auth.isCompleto();

  raiz.querySelectorAll('.somente-admin').forEach((el) => {
    if (!admin && el.dataset.ocultoPorPermissao !== '1') {
      el.dataset.ocultoPorPermissao = '1';
      el.style.display = 'none';
    } else if (admin && el.dataset.ocultoPorPermissao === '1') {
      delete el.dataset.ocultoPorPermissao;
      el.style.display = '';
    }
  });
}

// Reaplica as permissões em elementos inseridos depois (tabelas, modais, árvore renderizados via JS).
// Só observa childList: alterar style/dataset não dispara o observer novamente.
function observarPermissoes() {
  aplicarPermissoes();
  new MutationObserver(() => aplicarPermissoes()).observe(document.body, { childList: true, subtree: true });
}

// Preenche o <select> com a árvore de GET /api/categorias, indentando por nível.
// Opções: placeholder (texto da 1ª opção, valor ""), selecionado (id) e ignorarId (omite a categoria
// e suas subcategorias; útil ao escolher o pai na edição).
function preencherSelectCategorias(select, arvore, { placeholder = 'Selecione', selecionado = null, ignorarId = null } = {}) {
  select.replaceChildren(new Option(placeholder, ''));

  const percorrer = (nos, nivel) => {
    nos.forEach((no) => {
      if (no.id === ignorarId) return;
      // Espaços comuns colapsam dentro de <option>; por isso NBSP.
      const prefixo = nivel > 0 ? '\u00A0\u00A0\u00A0\u00A0'.repeat(nivel) + '└ ' : '';
      select.appendChild(new Option(prefixo + no.nome, String(no.id)));
      percorrer(no.filhos || [], nivel + 1);
    });
  };
  percorrer(arvore || [], 0);

  select.value = selecionado == null ? '' : String(selecionado);
  if (select.selectedIndex === -1) select.value = '';
}

function criarBotaoAcao(icone, titulo, classeCor, aoClicar) {
  const botao = document.createElement('button');
  botao.type = 'button';
  botao.title = titulo;
  botao.setAttribute('aria-label', titulo);
  botao.className = `rounded-lg p-1.5 text-slate-500 hover:bg-slate-100 ${classeCor}`;
  botao.innerHTML = svgIcon(icone);
  botao.addEventListener('click', aoClicar);
  return botao;
}

function criarListaCategorias(nos, nivel, { onEditar, onExcluir }) {
  const lista = document.createElement('ul');
  lista.className = nivel === 0 ? 'space-y-1' : 'ml-4 mt-1 space-y-1 border-l border-gray-100 pl-3';

  nos.forEach((no) => {
    const item = document.createElement('li');

    const linha = document.createElement('div');
    linha.className = 'flex items-center justify-between rounded-lg px-3 py-2 hover:bg-slate-50';

    const nome = document.createElement('span');
    nome.className = nivel === 0 ? 'text-sm font-medium text-slate-800' : 'text-sm text-slate-700';
    nome.textContent = no.nome;

    const acoes = document.createElement('div');
    acoes.className = 'somente-admin flex items-center gap-1';
    acoes.append(
      criarBotaoAcao('edit', 'Editar', 'hover:text-indigo-600', () => onEditar && onEditar(no)),
      criarBotaoAcao('trash', 'Excluir', 'hover:text-red-600', () => onExcluir && onExcluir(no)),
    );

    linha.append(nome, acoes);
    item.appendChild(linha);
    if (no.filhos && no.filhos.length > 0) {
      item.appendChild(criarListaCategorias(no.filhos, nivel + 1, { onEditar, onExcluir }));
    }
    lista.appendChild(item);
  });

  return lista;
}

// Desenha a árvore de categorias em container. Callbacks recebem a categoria clicada.
function renderizarArvoreCategorias(container, arvore, { onEditar, onExcluir } = {}) {
  container.replaceChildren();

  if (!arvore || arvore.length === 0) {
    const vazio = document.createElement('p');
    vazio.className = 'py-8 text-center text-sm text-slate-500';
    vazio.textContent = 'Nenhuma categoria cadastrada.';
    container.appendChild(vazio);
    return;
  }

  container.appendChild(criarListaCategorias(arvore, 0, { onEditar, onExcluir }));
  aplicarPermissoes(container);
}

// Liga o formulário de despesas: máscara no valor e campo de parcelas visível só para
// RECORRENTE e PARCELAMENTO. Os seletores aceitam ids diferentes dos padrões (nomes dos campos da API).
// O campo de parcelas é escondido junto do ancestral [data-campo] (ou do pai direto).
// Retorna atualizar(): chame depois de alterar o tipo por código ou de form.reset(), que não dispara "change".
function configurarCamposDinamicosDespesas({ tipo = '#tipo', valor = '#valor', parcelas = '#total_parcelas' } = {}) {
  const tipoEl = document.querySelector(tipo);
  const valorEl = document.querySelector(valor);
  const parcelasEl = document.querySelector(parcelas);
  if (!tipoEl) return () => {};

  if (valorEl && !valorEl.dataset.mascaraAtiva) {
    valorEl.dataset.mascaraAtiva = '1';
    valorEl.addEventListener('input', () => mascaraMoeda(valorEl));
  }

  const grupo = parcelasEl && (parcelasEl.closest('[data-campo]') || parcelasEl.parentElement);
  const rotulo = parcelasEl && parcelasEl.id ? document.querySelector(`label[for="${CSS.escape(parcelasEl.id)}"]`) : null;

  const atualizar = () => {
    if (!parcelasEl) return;

    const parcelado = tipoEl.value === 'RECORRENTE' || tipoEl.value === 'PARCELAMENTO';
    grupo.style.display = parcelado ? '' : 'none';
    parcelasEl.required = parcelado;
    parcelasEl.min = '2';
    parcelasEl.max = '120';
    if (!parcelado) parcelasEl.value = '';
    if (rotulo && parcelado) rotulo.textContent = tipoEl.value === 'RECORRENTE' ? 'Quantidade de meses' : 'Número de parcelas';
  };

  if (!tipoEl.dataset.dinamicoAtivo) {
    tipoEl.dataset.dinamicoAtivo = '1';
    tipoEl.addEventListener('change', atualizar);
  }
  atualizar();
  return atualizar;
}

// Cria um elemento; o texto entra via textContent (nunca interpretado como HTML).
function criarEl(tag, classe, texto) {
  const el = document.createElement(tag);
  if (classe) el.className = classe;
  if (texto != null) el.textContent = texto;
  return el;
}

function hojeISO() {
  const h = new Date();
  return `${h.getFullYear()}-${String(h.getMonth() + 1).padStart(2, '0')}-${String(h.getDate()).padStart(2, '0')}`;
}

// Período padrão dos filtros: do 1º dia do mês atual até 31/12 do ano.
function periodoPadrao() {
  const h = new Date();
  return { inicial: `${h.getFullYear()}-${String(h.getMonth() + 1).padStart(2, '0')}-01`, final: `${h.getFullYear()}-12-31` };
}

// Modais: o elemento raiz leva data-modal; botões com data-fechar-modal fecham. Esc e clique no fundo também.
function abrirModal(id) {
  const modal = document.getElementById(id);
  modal.classList.remove('hidden');
  modal.classList.add('flex');
  document.body.classList.add('overflow-hidden');
  const primeiro = modal.querySelector('input:not([disabled]), select:not([disabled]), textarea:not([disabled])');
  if (primeiro) primeiro.focus();
}

function fecharModal(id) {
  const modal = document.getElementById(id);
  modal.classList.add('hidden');
  modal.classList.remove('flex');
  if (!document.querySelector('[data-modal].flex')) document.body.classList.remove('overflow-hidden');
}

(() => {
  // O clique só fecha se também começou no fundo: arrastar uma seleção de texto para fora não fecha.
  let origem = null;
  document.addEventListener('mousedown', (e) => { origem = e.target; });
  document.addEventListener('click', (e) => {
    const botao = e.target.closest('[data-fechar-modal]');
    if (botao) {
      fecharModal(botao.closest('[data-modal]').id);
    } else if (e.target.hasAttribute('data-modal') && origem === e.target) {
      fecharModal(e.target.id);
    }
  });
  document.addEventListener('keydown', (e) => {
    if (e.key !== 'Escape') return;
    const abertos = document.querySelectorAll('[data-modal].flex');
    if (abertos.length) fecharModal(abertos[abertos.length - 1].id);
  });
})();

function mostrarToast(mensagem, tipo = 'sucesso') {
  let area = document.getElementById('area-toasts');
  if (!area) {
    area = criarEl('div', 'fixed right-4 top-4 z-[60] flex w-80 max-w-[calc(100vw-2rem)] flex-col gap-2');
    area.id = 'area-toasts';
    document.body.appendChild(area);
  }

  const cor = tipo === 'erro' ? 'border-red-200 bg-red-50 text-red-700' : 'border-emerald-200 bg-emerald-50 text-emerald-700';
  const toast = criarEl('div', `animar-aparecer rounded-lg border px-4 py-3 text-sm shadow-sm ${cor}`, mensagem);
  toast.setAttribute('role', tipo === 'erro' ? 'alert' : 'status');
  area.appendChild(toast);
  setTimeout(() => toast.remove(), 4500);
}

const BADGES_TIPO = {
  PADRAO: ['Padrão', 'bg-blue-100 text-blue-700'],
  EVENTUAL: ['Eventual', 'bg-amber-100 text-amber-700'],
  RECORRENTE: ['Recorrente', 'bg-violet-100 text-violet-700'],
  PARCELAMENTO: ['Parcelamento', 'bg-emerald-100 text-emerald-700'],
};

function criarBadgeTipo(tipo) {
  const [rotulo, cor] = BADGES_TIPO[tipo] || [tipo, 'bg-gray-100 text-gray-600'];
  return criarEl('span', `inline-block rounded-full px-2 py-0.5 text-xs font-medium ${cor}`, rotulo);
}

function criarBadgeNivel(nivel) {
  return criarEl('span', `inline-block rounded-full px-2 py-0.5 text-xs font-medium ${CLASSE_NIVEL[nivel] || CLASSE_NIVEL.READ}`, nivel);
}

// Rodapé de tabela: "Página X de Y" e botões; aoMudar recebe o número da nova página.
function renderizarPaginacao(container, { pagina, totalPaginas, totalRegistros }, aoMudar) {
  container.replaceChildren();

  const info = criarEl('span', 'text-sm text-slate-500',
    `Página ${pagina} de ${totalPaginas} · ${totalRegistros} registro${totalRegistros === 1 ? '' : 's'}`);

  const anterior = criarEl('button', 'botao botao-secundario', 'Anterior');
  anterior.type = 'button';
  anterior.disabled = pagina <= 1;
  anterior.addEventListener('click', () => aoMudar(pagina - 1));

  const proxima = criarEl('button', 'botao botao-secundario', 'Próxima');
  proxima.type = 'button';
  proxima.disabled = pagina >= totalPaginas;
  proxima.addEventListener('click', () => aoMudar(pagina + 1));

  const nav = criarEl('div', 'flex gap-2');
  nav.append(anterior, proxima);
  container.append(info, nav);
}

// Diálogo de escolha. opcoes: [{ rotulo, valor, classe }]; "Cancelar" é sempre incluído e resolve null.
function escolherAcao({ titulo, mensagem, opcoes }) {
  return new Promise((resolve) => {
    const fundo = criarEl('div', 'fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4');
    const caixa = criarEl('div', 'animar-aparecer w-full max-w-md rounded-xl border border-gray-100 bg-white p-6 shadow-sm');
    caixa.setAttribute('role', 'alertdialog');
    caixa.setAttribute('aria-modal', 'true');
    const botoes = criarEl('div', 'mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end');

    const aoTecla = (e) => {
      if (e.key !== 'Escape') return;
      e.stopPropagation();
      fechar(null);
    };
    const fechar = (valor) => {
      document.removeEventListener('keydown', aoTecla, true);
      fundo.remove();
      if (!document.querySelector('[data-modal].flex')) document.body.classList.remove('overflow-hidden');
      resolve(valor);
    };

    [{ rotulo: 'Cancelar', valor: null, classe: 'botao-secundario' }, ...opcoes].forEach((opcao) => {
      const botao = criarEl('button', `botao ${opcao.classe}`, opcao.rotulo);
      botao.type = 'button';
      botao.addEventListener('click', () => fechar(opcao.valor));
      botoes.appendChild(botao);
    });

    fundo.addEventListener('click', (e) => { if (e.target === fundo) fechar(null); });
    caixa.append(criarEl('h2', 'text-lg font-semibold text-slate-800', titulo), criarEl('p', 'mt-2 text-sm text-slate-600', mensagem), botoes);
    fundo.appendChild(caixa);
    document.body.appendChild(fundo);
    document.body.classList.add('overflow-hidden');
    document.addEventListener('keydown', aoTecla, true);
    botoes.firstChild.focus();
  });
}

function confirmarAcao({ titulo, mensagem, rotulo = 'Excluir' }) {
  return escolherAcao({ titulo, mensagem, opcoes: [{ rotulo, valor: true, classe: 'botao-perigo' }] })
    .then((valor) => valor === true);
}

// Registrado depois de layout.js, então roda após o layout ser montado.
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', observarPermissoes);
} else {
  observarPermissoes();
}
