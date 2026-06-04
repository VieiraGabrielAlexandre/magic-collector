# Frontend Responsive Spec — Magic Collector

## Visão

O frontend do Magic Collector deve oferecer uma experiência mobile-first, elegante e funcional para gerenciamento de coleção, decks e partidas de Magic: The Gathering.

A interface deve funcionar bem em celulares pequenos, tablets/iPads, desktops e telas grandes, sem comprometer nenhuma funcionalidade existente.

---

## Escopo

Esta especificação cobre apenas o frontend:

* `frontend/src/App.jsx`
* `frontend/src/App.css`
* `frontend/src/services/api.js`, somente se necessário para ajustes de payload já existentes

Não faz parte do escopo:

* alterar backend
* alterar banco de dados
* alterar rotas da API
* alterar contratos existentes
* migrar para TypeScript
* adicionar roteamento
* adicionar Redux
* reescrever a aplicação inteira

---

## Princípios

* Mobile-first.
* Interface funcional antes de visualmente bonita.
* Visual premium sem poluição.
* Responsividade real, não apenas media query pontual.
* Botões sempre clicáveis/toqueáveis.
* Conteúdo importante sempre visível.
* Acessibilidade básica obrigatória.
* Código simples e compatível com JSX puro.
* Mudanças incrementais e seguras.

---

## Dispositivos-alvo

A UI deve ser validada nos seguintes tamanhos:

| Categoria      | Viewport  |
| -------------- | --------- |
| Mobile pequeno | 320x740   |
| Mobile comum   | 390x844   |
| Mobile grande  | 430x932   |
| iPad           | 768x1024  |
| iPad Air       | 820x1180  |
| iPad Pro       | 1024x1366 |
| Desktop        | 1440x900  |
| Wide           | 1920x1080 |

---

## Breakpoints

Usar como referência:

```css
/* base: mobile first */
/* sm: 640px */
/* md: 768px */
/* lg: 1024px */
/* xl: 1280px */
/* 2xl: 1536px */
```

Regras:

* Começar pelo layout mobile.
* Expandir progressivamente para tablets e desktops.
* Evitar CSS específico demais para um único aparelho.
* Preferir `grid`, `flex-wrap`, `minmax()`, `clamp()`, `max-width` e `width: 100%`.

---

## Regras globais obrigatórias

* Não pode haver scroll horizontal acidental.
* Nenhum botão pode ficar sobreposto.
* Nenhum botão pode sair da tela.
* Nenhum texto importante pode ficar cortado sem alternativa.
* Modais precisam caber em 320px.
* Cards precisam se adaptar sem quebrar layout.
* Imagens precisam preservar proporção.
* Ações principais precisam continuar acessíveis no mobile.
* Funcionalidade essencial não pode depender de hover.
* Inputs e botões devem ser confortáveis para toque.
* Telas grandes não devem esticar conteúdo indefinidamente.

---

## Área mínima de toque

Elementos interativos devem respeitar:

* Botões: mínimo 44px de altura.
* Ícones clicáveis: área mínima de 44x44px.
* Inputs/selects: preferencialmente 44px ou mais.
* Espaçamento suficiente entre ações destrutivas e ações principais.

---

## Layout global

O layout principal deve:

* usar container centralizado;
* ter `max-width` em desktop;
* ter padding menor em mobile e maior em desktop;
* evitar largura fixa;
* usar espaçamento consistente;
* respeitar safe area quando aplicável;
* impedir overflow horizontal.

Exemplo conceitual:

```css
.app-shell {
  width: 100%;
  max-width: 1440px;
  margin: 0 auto;
  padding-inline: clamp(1rem, 3vw, 2rem);
}
```

---

## Header

O header deve:

* ser compacto em mobile;
* manter título e ações principais legíveis;
* evitar botões espremidos;
* permitir quebra ou empilhamento em telas pequenas;
* melhorar presença visual em desktop;
* manter hierarquia clara entre título, subtítulo e ações.

Em mobile:

* título em tamanho responsivo;
* ações empilhadas ou organizadas em linha quebrável;
* nenhum elemento sobreposto.

Em desktop:

* header pode usar layout horizontal;
* ações podem ficar à direita;
* conteúdo deve permanecer dentro de container com `max-width`.

---

## Navegação por abas

As abas principais provavelmente representam:

* coleção
* decks
* batalhas

A navegação deve:

* ter estado ativo claro;
* funcionar bem com toque;
* não quebrar em 320px;
* não esconder funcionalidades essenciais;
* permitir scroll horizontal controlado se necessário;
* manter foco visível para teclado.

Em mobile:

* abas podem ocupar largura total;
* podem usar scroll horizontal discreto;
* cada aba deve ter área mínima de toque.

Em desktop:

* abas devem parecer parte da estrutura principal da aplicação;
* evitar aparência de botões soltos.

---

## Coleção de cartas

A tela de coleção deve priorizar:

* busca;
* filtros;
* listagem;
* ações rápidas;
* acesso ao detalhe da carta.

Cada carta deve apresentar, quando disponível:

* imagem;
* nome;
* nome impresso/localizado;
* set;
* número de coleção;
* raridade;
* cor/identidade;
* condição;
* flags como foil, prerelease, commander;
* deck associado;
* ações principais.

Em mobile:

* layout de uma coluna ou grid seguro;
* imagem não pode deformar;
* ações devem empilhar ou ficar em menu compacto;
* textos longos devem quebrar corretamente;
* badges devem fazer wrap.

Em tablet:

* usar grid com 2 colunas quando fizer sentido;
* aproveitar melhor o espaço sem perder legibilidade.

Em desktop:

* grid com múltiplas colunas;
* cards com largura máxima saudável;
* evitar cards muito largos ou muito altos sem necessidade.

---

## Cards

Cards devem ter:

* background de superfície;
* borda sutil;
* raio consistente;
* sombra discreta;
* espaçamento interno confortável;
* hierarquia visual clara.

Estados:

* hover discreto em desktop;
* foco visível;
* estado selecionado, se existir;
* estado disabled, se existir.

Não fazer:

* não depender apenas de hover;
* não usar sombras exageradas;
* não usar gradientes que prejudiquem leitura;
* não colocar muitos botões lado a lado no mobile.

---

## Imagens de cartas

As imagens devem:

* preservar proporção original;
* usar `max-width: 100%`;
* usar `height: auto`;
* ter fallback se ausentes;
* não estourar o card;
* carregar de forma visualmente estável.

Para cartas double-faced:

* se já houver suporte, manter funcionando;
* se não houver, não quebrar layout quando `card_faces` existir.

---

## Formulários

Formulários devem:

* ser legíveis em mobile;
* usar labels claros;
* ter inputs confortáveis;
* quebrar em uma coluna no mobile;
* usar duas ou mais colunas em tablet/desktop quando fizer sentido;
* manter botões de salvar/cancelar visíveis;
* exibir erros de forma clara;
* não ultrapassar a largura da tela.

Campos comuns:

* nome;
* set code;
* número de coleção;
* idioma;
* raridade;
* condição;
* flags;
* deck;
* quantidade, se existir.

---

## Busca, filtros e ordenação

Busca e filtros devem:

* ser acessíveis no mobile;
* não ocupar espaço excessivo;
* não quebrar layout;
* permitir limpar filtros, se já houver essa ação;
* manter estado visual claro;
* funcionar sem scroll horizontal.

Em mobile:

* filtros podem ficar empilhados;
* filtros avançados podem ficar colapsados ou em painel;
* botões de aplicar/limpar devem ser fáceis de tocar.

Em desktop:

* filtros podem ficar em linha ou grid;
* buscar bom equilíbrio entre densidade e legibilidade.

---

## Decks

A tela de decks deve destacar:

* nome do deck;
* cores;
* ícone do set;
* quantidade de cartas;
* tema visual;
* ações principais.

Para Commander, quando aplicável:

* destacar comandante;
* destacar identidade de cor;
* alertar visualmente caso regras básicas não estejam respeitadas, se essa informação já existir no frontend.

Ações de deck:

* editar;
* excluir;
* importar precon;
* importar lista;
* atualizar ícone.

Em mobile:

* ações não devem ficar espremidas;
* cards de deck devem empilhar;
* botões destrutivos devem ficar separados visualmente.

---

## Battles

A tela de battles deve mostrar claramente:

* resultado;
* formato;
* deck usado;
* oponente;
* data;
* observações.

Resultados devem ter destaque:

* vitória;
* derrota;
* empate.

Em mobile:

* preferir cards em vez de tabela larga;
* dados devem quebrar linha corretamente;
* ações devem ficar acessíveis.

Em desktop:

* pode usar layout mais denso;
* histórico deve ser fácil de escanear.

---

## Modais

Modais devem:

* caber em 320px;
* ter `max-width`;
* ter `max-height`;
* permitir scroll interno;
* manter botão de fechar acessível;
* manter ações finais visíveis;
* não cortar conteúdo importante.

Em mobile:

* modal pode parecer bottom sheet ou ocupar quase tela inteira;
* padding menor;
* botões podem empilhar.

Em desktop:

* modal centralizado;
* largura confortável;
* conteúdo bem distribuído.

---

## Estados de interface

A UI deve tratar visualmente:

* loading;
* erro;
* vazio;
* sucesso;
* disabled;
* foco;
* hover;
* active.

Estados vazios devem ser úteis, por exemplo:

* nenhuma carta encontrada;
* nenhum deck cadastrado;
* nenhuma batalha registrada;
* busca sem resultado.

---

## Identidade visual MTG

A estética deve ser inspirada em Magic: The Gathering de forma sutil.

Direção visual:

* dark fantasy moderno;
* dourado/cobre como destaque;
* grafite para superfícies;
* bordas premium;
* sombras suaves;
* raridade com cores diferenciadas;
* badges de cor WUBRG;
* detalhes que lembrem card game sem copiar assets oficiais.

Evitar:

* excesso de neon;
* fundos muito carregados;
* efeitos que prejudiquem leitura;
* visual infantil;
* muitas cores competindo ao mesmo tempo.

---

## Paleta sugerida

Usar ou adaptar os tokens existentes.

```css
:root {
  --bg: #0f0d0a;
  --surface: #18140f;
  --surface-2: #211b14;

  --border: rgba(212, 175, 55, 0.22);
  --border-strong: rgba(212, 175, 55, 0.42);

  --primary: #d4af37;
  --primary-soft: #f1d884;
  --secondary: #6d5dfc;

  --text: #f4efe3;
  --text-muted: #b9ad98;

  --success: #4f9d69;
  --danger: #b84a3a;
  --warning: #d9902f;

  --mana-white: #f8f0d8;
  --mana-blue: #4f8cc9;
  --mana-black: #3a3332;
  --mana-red: #c44f3d;
  --mana-green: #3f8f5f;
  --mana-colorless: #aaa39a;
}
```

Ajustar conforme contraste real da aplicação.

---

## Tipografia

A tipografia deve:

* ser legível;
* ter escala fluida;
* diferenciar títulos, subtítulos e texto comum;
* evitar textos pequenos demais em mobile.

Sugestão:

```css
.page-title {
  font-size: clamp(1.5rem, 4vw, 2.5rem);
}

.section-title {
  font-size: clamp(1.125rem, 2.5vw, 1.5rem);
}

.body-text {
  font-size: clamp(0.875rem, 1.6vw, 1rem);
}
```

---

## Espaçamento

Usar escala consistente:

* 4px
* 8px
* 12px
* 16px
* 24px
* 32px
* 48px

Evitar margens aleatórias e correções pontuais com valores mágicos.

---

## Acessibilidade

Obrigatório:

* foco visível;
* labels em inputs;
* `aria-label` em botões de ícone;
* contraste adequado;
* não comunicar informação apenas por cor;
* textos alternativos em imagens;
* botões com nomes acessíveis;
* modais com fechamento claro;
* navegação básica por teclado.

---

## Performance visual

Evitar:

* animações pesadas;
* blur excessivo;
* sombras muito caras;
* imagens sem limite de tamanho;
* re-renderizações desnecessárias causadas por refatoração visual.

Permitido:

* transições simples;
* hover leve;
* foco claro;
* skeleton/loading simples, se já for compatível com o projeto.

---

## Código

Como o frontend está concentrado em `App.jsx` e `App.css`:

* preferir extrair pequenos componentes internos quando melhorar legibilidade;
* não criar arquitetura complexa;
* não introduzir gerenciamento de estado novo;
* manter nomes de classes claros;
* evitar estilos inline;
* remover duplicações óbvias;
* preservar fluxo atual de estado.

Exemplos de componentes internos aceitáveis:

* `ResponsiveTabs`
* `CardGrid`
* `ActionGroup`
* `EmptyState`
* `ModalShell`
* `ManaBadge`
* `RarityBadge`

A extração só deve ser feita se reduzir complexidade.

---

## Critérios de aceite finais

A implementação atende esta spec quando:

* funciona sem overflow horizontal em 320px;
* todos os botões seguem área mínima de toque;
* header e tabs funcionam em mobile;
* coleção é legível em mobile, tablet e desktop;
* decks são legíveis e visualmente coerentes;
* battles são fáceis de ler;
* modais cabem em telas pequenas;
* ações destrutivas são distinguíveis;
* a UI mantém identidade dark/gold MTG;
* acessibilidade básica é respeitada;
* não houve alteração no backend;
* não houve alteração em contratos da API;
* o projeto continua simples.

---

## Checklist manual final

Validar:

* abrir aplicação em 320px;
* navegar pelas abas;
* buscar carta;
* adicionar carta;
* editar carta;
* excluir carta;
* abrir detalhe da carta;
* mover carta para deck;
* criar/editar/excluir deck;
* importar precon;
* importar lista;
* atualizar ícone do deck;
* registrar battle;
* excluir battle;
* abrir e fechar todos os modais;
* validar foco via teclado;
* validar ausência de scroll horizontal;
* validar desktop em 1440px;
* validar wide em 1920px.
