# Frontend Responsive Plan — Magic Collector

## Objetivo

Refatorar o frontend do Magic Collector para ficar altamente responsivo, elegante e funcional em mobile, tablets/iPads, desktops e telas grandes, mantendo coerência visual com Magic: The Gathering.

Este plano deve ser executado por fases. Não implementar tudo de uma vez.

---

## Regras globais

* Não alterar backend.
* Não alterar contratos da API.
* Não alterar regras de negócio.
* Não adicionar bibliotecas sem justificativa.
* Não migrar para TypeScript.
* Não quebrar o fluxo atual baseado em `activeTab`.
* Manter React 18 + Vite + JSX puro.
* Priorizar mudanças em `frontend/src/App.jsx` e `frontend/src/App.css`.
* Evitar refatoração grande demais se pequenos componentes internos resolverem.
* Garantir que a UI funcione de 320px até 1920px+.
* Nenhum botão pode ficar sobreposto.
* Nenhum conteúdo importante pode ficar cortado.
* Não pode haver scroll horizontal acidental.
* Funcionalidades essenciais não podem depender de hover.

---

## Breakpoints obrigatórios

Validar visualmente em:

| Nome           | Tamanho   |
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

## Critérios gerais de aceite

A refatoração só estará concluída quando:

* Header funcionar bem em mobile e desktop.
* Tabs não quebrarem nem ficarem sobrepostas.
* Cards de cartas ficarem legíveis em mobile.
* Modais couberem na tela.
* Formulários forem confortáveis no toque.
* Botões tiverem área mínima de toque.
* Listagens não gerarem overflow horizontal.
* Grids se adaptarem bem a mobile, tablet e desktop.
* Telas grandes não ficarem esticadas demais.
* Visual geral parecer mais premium e coerente com Magic: The Gathering.

---

# Fase 0 — Diagnóstico

## Objetivo

Entender a estrutura atual antes de codificar.

## Tarefas

* Ler `frontend/src/App.jsx`.
* Ler `frontend/src/App.css`.
* Identificar os principais blocos visuais:

  * header
  * tabs
  * coleção
  * decks
  * battles
  * formulários
  * modais
  * cards
  * filtros
  * botões
* Identificar pontos de quebra em mobile/tablet.
* Identificar classes CSS reutilizáveis.
* Identificar estilos duplicados ou muito rígidos.

## Não fazer

* Não alterar código ainda.

## Entrega esperada

Criar um resumo com:

* problemas encontrados
* arquivos impactados
* componentes/blocos prioritários
* riscos de regressão
* ordem sugerida de execução

---

# Fase 1 — Fundação responsiva global

## Objetivo

Criar uma base sólida para responsividade em toda a aplicação.

## Tarefas

* Garantir `box-sizing: border-box`.
* Garantir que `html`, `body` e `#root` não causem overflow horizontal.
* Revisar containers principais.
* Criar ou ajustar classes globais de layout.
* Definir `max-width` para telas grandes.
* Usar `width: 100%` onde fizer sentido.
* Remover larguras fixas problemáticas.
* Ajustar espaçamento global com escala consistente.
* Garantir que imagens nunca estourem o container.

## Critérios de aceite

* Aplicação não deve gerar scroll horizontal em 320px.
* Conteúdo principal deve respirar bem em desktop.
* Base visual deve continuar dark/gold.
* Nenhuma funcionalidade deve ser removida.

---

# Fase 2 — Header, navegação e tabs

## Objetivo

Garantir que navegação e ações principais funcionem perfeitamente em qualquer tela.

## Tarefas

* Refatorar header para mobile-first.
* Ajustar título, subtítulo e ações do topo.
* Ajustar tabs `collection`, `decks` e `battles`.
* Em mobile, permitir quebra elegante ou scroll horizontal controlado nas tabs.
* Garantir área de toque mínima de 44px.
* Evitar botões espremidos.
* Garantir estado ativo claro.
* Melhorar visual desktop sem esticar demais.

## Critérios de aceite

* Tabs acessíveis em 320px.
* Nenhum botão sobreposto.
* Estado ativo evidente.
* Header elegante em desktop.
* Header compacto em mobile.

---

# Fase 3 — Formulários e filtros

## Objetivo

Deixar formulários e filtros confortáveis em mobile e eficientes em desktop.

## Tarefas

* Revisar campos de cadastro/edição de carta.
* Revisar busca, filtros, ordenação e paginação.
* Inputs devem ter altura confortável.
* Labels devem ser legíveis.
* Grupos de campos devem quebrar em uma coluna no mobile.
* Em tablet/desktop, usar grid com 2 ou mais colunas quando fizer sentido.
* Botões de ação devem ficar claros e não sobrepostos.
* Selects e inputs não podem ultrapassar largura da tela.

## Critérios de aceite

* Formulários usáveis com toque.
* Inputs não cortam texto importante.
* Filtros não quebram layout.
* Botões principais continuam visíveis.
* Mobile usa uma coluna.
* Tablet/desktop aproveitam melhor o espaço.

---

# Fase 4 — Cards da coleção

## Objetivo

Melhorar listagem e visualização das cartas da coleção.

## Tarefas

* Ajustar grid/lista de cartas.
* Garantir imagens com proporção correta.
* Em mobile, cards devem ocupar largura total ou grid seguro.
* Em tablet, usar grid intermediário.
* Em desktop, usar grid mais amplo sem esticar cards demais.
* Melhorar hierarquia visual:

  * nome da carta
  * set
  * raridade
  * cores
  * condição
  * deck
  * preço/quantidade se existir
* Organizar ações do card:

  * ver detalhes
  * editar
  * remover
  * mover para deck
* Evitar excesso de botões lado a lado no mobile.

## Critérios de aceite

* Cards legíveis em 320px.
* Imagens não deformam.
* Botões não se sobrepõem.
* Cards ficam elegantes em desktop.
* Visual remete mais a Magic sem poluir.

---

# Fase 5 — Decks

## Objetivo

Melhorar tela de decks para ficar mais clara, visual e coerente com MTG.

## Tarefas

* Ajustar listagem de decks.
* Destacar:

  * nome do deck
  * cores
  * set/icon
  * quantidade de cartas
  * theme color
* Melhorar cards de deck.
* Garantir que ações de deck não quebrem no mobile:

  * editar
  * deletar
  * importar precon
  * importar lista
  * atualizar ícone
* Pensar em visual de “deck box” ou “card game dashboard” de forma sutil.

## Critérios de aceite

* Tela de decks usável em mobile.
* Ações não ficam espremidas.
* Cards de deck têm boa leitura.
* Tablet e desktop aproveitam melhor o espaço.

---

# Fase 6 — Battles

## Objetivo

Melhorar a tela de histórico de partidas.

## Tarefas

* Ajustar formulário de batalha.
* Ajustar listagem/histórico.
* Destacar resultado:

  * vitória
  * derrota
  * empate
* Melhorar leitura de:

  * formato
  * deck usado
  * oponente
  * observações
  * data
* Em mobile, evitar tabela larga se existir.
* Transformar em cards responsivos, se necessário.

## Critérios de aceite

* Histórico legível em mobile.
* Resultados têm destaque visual claro.
* Nenhum dado importante fica cortado.
* Ações continuam acessíveis.

---

# Fase 7 — Modais e detalhes

## Objetivo

Garantir que todos os modais funcionem bem em telas pequenas.

## Tarefas

* Revisar modais existentes.
* Garantir `max-height` seguro.
* Permitir scroll interno quando necessário.
* Ajustar padding para mobile.
* Botões do modal devem empilhar em mobile se necessário.
* Detalhes da carta devem exibir bem:

  * imagem
  * nome
  * texto
  * tipo
  * raridade
  * set
  * dados externos da Scryfall
* Double-faced cards devem continuar funcionando se já houver suporte.

## Critérios de aceite

* Modal cabe em 320px.
* Modal não corta botões.
* Conteúdo longo pode rolar.
* Ações de fechar/salvar sempre acessíveis.

---

# Fase 8 — Identidade visual Magic: The Gathering

## Objetivo

Refinar a aparência geral sem comprometer usabilidade.

## Tarefas

* Revisar paleta dark/gold atual.
* Melhorar tokens CSS se necessário:

  * background
  * surface
  * border
  * primary
  * secondary
  * danger
  * success
  * warning
  * text-primary
  * text-muted
* Usar detalhes sutis de fantasia/card game:

  * bordas premium
  * sombras suaves
  * gradientes discretos
  * destaques dourados/cobre
  * raridade com cor coerente
* Melhorar estados:

  * hover
  * focus
  * active
  * disabled
  * loading
  * empty
  * error
* Evitar excesso de brilho, neon ou efeitos pesados.

## Critérios de aceite

* UI fica mais premium.
* Visual continua legível.
* Contraste permanece bom.
* Estilo combina com coleção de Magic.
* Não parece poluído.

---

# Fase 9 — Acessibilidade

## Objetivo

Garantir boa usabilidade com teclado, leitores de tela e contraste.

## Tarefas

* Garantir foco visível.
* Botões apenas com ícone precisam de `aria-label`.
* Inputs precisam de label ou descrição clara.
* Imagens importantes precisam de `alt`.
* Estados de erro devem ser textuais e visuais.
* Não usar apenas cor para comunicar informação importante.
* Garantir contraste adequado no tema dark.

## Critérios de aceite

* Navegação por teclado básica funciona.
* Foco visível em botões, inputs e tabs.
* Ícones acionáveis têm descrição.
* Informações críticas não dependem apenas de cor.

---

# Fase 10 — Validação manual

## Objetivo

Validar a aplicação após as mudanças.

## Checklist por viewport

Testar em:

* 320x740
* 390x844
* 430x932
* 768x1024
* 820x1180
* 1024x1366
* 1440x900
* 1920x1080

## Verificar

* Sem scroll horizontal acidental.
* Header não quebra.
* Tabs funcionam.
* Coleção carrega corretamente.
* Busca funciona.
* Filtros funcionam.
* Paginação funciona.
* Cadastro de carta funciona.
* Edição de carta funciona.
* Exclusão de carta funciona.
* Modal de detalhes funciona.
* Decks funcionam.
* Importação de deck continua acessível.
* Battles funcionam.
* Botões não ficam sobrepostos.
* Textos não ficam ilegíveis.
* Imagens não deformam.

---

# Fase 11 — Opcional: Harness com Playwright

## Objetivo

Criar validação automatizada mínima de responsividade.

## Importante

Só implementar se for simples e não adicionar complexidade excessiva.

## Tarefas

* Avaliar se vale adicionar Playwright.
* Criar teste básico de smoke visual.
* Testar principais viewports.
* Validar ausência de overflow horizontal.
* Validar que botões principais estão visíveis.

## Viewports sugeridas

```js
const viewports = [
  { name: "mobile-small", width: 320, height: 740 },
  { name: "mobile", width: 390, height: 844 },
  { name: "mobile-large", width: 430, height: 932 },
  { name: "ipad", width: 768, height: 1024 },
  { name: "ipad-air", width: 820, height: 1180 },
  { name: "ipad-pro", width: 1024, height: 1366 },
  { name: "desktop", width: 1440, height: 900 },
  { name: "wide", width: 1920, height: 1080 }
];
```

## Critérios de aceite

* Teste roda localmente.
* Teste não depende de dados externos frágeis.
* Teste detecta overflow horizontal.
* Teste valida carregamento básico da UI.

--- 