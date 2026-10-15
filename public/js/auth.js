'use strict';

// BASE_PATH é o prefixo da aplicação: "/app" para um script em "/app/js/auth.js", "" na raiz.
(() => {
  const caminho = document.currentScript ? new URL(document.currentScript.src, location.href).pathname : '';
  const m = caminho.match(/^(.*)\/js\/[^/]*$/);
  window.BASE_PATH = m ? m[1] : '';
})();

function irParaLogin() {
  location.replace(`${window.BASE_PATH}/login.html`);
}

const Auth = {
  getToken() {
    return sessionStorage.getItem('token');
  },

  getUser() {
    return {
      nome: sessionStorage.getItem('nome') || '',
      email: sessionStorage.getItem('email') || '',
      nivel: sessionStorage.getItem('nivel') || '',
    };
  },

  // Lê o claim id do JWT só para decisões de interface; quem autoriza é sempre o servidor.
  getUserId() {
    try {
      const payload = Auth.getToken().split('.')[1].replace(/-/g, '+').replace(/_/g, '/');
      return JSON.parse(atob(payload)).id ?? null;
    } catch {
      return null;
    }
  },

  // Recebe a resposta de POST /api/auth/login.
  setSession({ token, nome, email, nivel }) {
    sessionStorage.setItem('token', token);
    sessionStorage.setItem('nome', nome);
    sessionStorage.setItem('email', email);
    sessionStorage.setItem('nivel', nivel);
  },

  // Acesso completo (ADMIN): pode incluir, editar e excluir.
  isCompleto() {
    return sessionStorage.getItem('nivel') === 'ADMIN';
  },

  isRead() {
    return sessionStorage.getItem('nivel') === 'READ';
  },

  logout() {
    ['token', 'nome', 'email', 'nivel'].forEach((chave) => sessionStorage.removeItem(chave));
    irParaLogin();
  },
};

// Protege a rota: qualquer página, exceto o login, exige token.
(() => {
  const emLogin = /\/login\.html$/.test(location.pathname);
  if (emLogin) return;

  if (!Auth.getToken()) irParaLogin();

  // Voltar após o logout pode reexibir a página do cache do navegador.
  window.addEventListener('pageshow', (evento) => {
    if (evento.persisted && !Auth.getToken()) irParaLogin();
  });
})();
