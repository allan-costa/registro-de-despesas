'use strict';

const FORMATADOR_BRL = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' });

// Aceita number ou string numérica ("1234.56"); valor inválido vira R$ 0,00.
function formatarMoeda(valor) {
  const n = Number(valor);
  return FORMATADOR_BRL.format(Number.isFinite(n) ? n : 0);
}

// Lê o dia direto da string: new Date() deslocaria a data conforme o fuso do navegador.
function formatarData(data) {
  if (!data) return '';

  if (data instanceof Date) {
    const dia = String(data.getDate()).padStart(2, '0');
    const mes = String(data.getMonth() + 1).padStart(2, '0');
    return `${dia}/${mes}/${data.getFullYear()}`;
  }

  const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(String(data));
  return m ? `${m[3]}/${m[2]}/${m[1]}` : '';
}

// Máscara de centavos: "1" -> "0,01" e "123456" -> "1.234,56" (máx. 12 dígitos, limite do NUMERIC(12,2)).
// Para preencher o campo com um número, atribua Math.round(valor * 100) ao value e chame a função.
// Aceita o elemento (oninput="mascaraMoeda(this)") ou o evento.
function mascaraMoeda(alvo) {
  const input = alvo.target || alvo;
  const digitos = input.value.replace(/\D/g, '').replace(/^0+/, '').slice(0, 12);
  if (!digitos) {
    input.value = '';
    return;
  }

  const centavos = digitos.padStart(3, '0');
  const inteiro = centavos.slice(0, -2).replace(/\B(?=(\d{3})+(?!\d))/g, '.');
  input.value = `${inteiro},${centavos.slice(-2)}`;
}

// Converte "1.234,56" (ou "R$ 1.234,56") em 1234.56; vazio ou inválido vira 0.
function moedaParaFloat(texto) {
  const normalizado = String(texto ?? '').replace(/[^\d,]/g, '').replace(',', '.');
  const n = parseFloat(normalizado);
  return Number.isFinite(n) ? n : 0;
}

class ApiError extends Error {
  constructor(mensagem, status) {
    super(mensagem);
    this.status = status;
  }
}

// Chama a API com o token (X-Auth-Token) e devolve o JSON (null se não houver corpo).
// 401 encerra a sessão, exceto com publico: true (login, onde 401 é senha errada).
// Erros viram ApiError com a mensagem do servidor.
async function apiFetch(caminho, { method = 'GET', body, publico = false } = {}) {
  const headers = {};
  if (!publico) headers['X-Auth-Token'] = Auth.getToken() || '';
  if (body !== undefined) headers['Content-Type'] = 'application/json';

  let resposta;
  try {
    resposta = await fetch(`${window.BASE_PATH}${caminho}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    throw new ApiError('Não foi possível conectar ao servidor.', 0);
  }

  if (resposta.status === 401 && !publico) {
    Auth.logout();
    throw new ApiError('Sessão expirada. Entre novamente.', 401);
  }

  let dados = null;
  try {
    dados = await resposta.json();
  } catch {
    // corpo vazio ou não JSON
  }
  if (!resposta.ok) throw new ApiError((dados && dados.erro) || `Erro ${resposta.status}`, resposta.status);
  return dados;
}

// Monta "?a=1&b=2" ignorando valores vazios.
function montarQuery(params) {
  const q = new URLSearchParams();
  Object.entries(params).forEach(([chave, valor]) => {
    if (valor !== '' && valor != null) q.set(chave, valor);
  });
  const texto = q.toString();
  return texto ? `?${texto}` : '';
}
