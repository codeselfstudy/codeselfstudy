" Some settings to make Vim nicer
set nocompatible
set number
set syntax
set history=999
set nomodeline
set list lcs=tab:>-,trail:\.,extends:»,precedes:«,nbsp:%
set showmatch
set matchtime=4
set mouse=a
set lazyredraw
set smarttab
set softtabstop=4
set shiftwidth=4
set tabstop=4
set expandtab
autocmd FileType ruby setlocal shiftwidth=2 tabstop=2 softtabstop=2
autocmd FileType eruby setlocal shiftwidth=2 tabstop=2 softtabstop=2
autocmd FileType jade setlocal shiftwidth=2 tabstop=2 softtabstop=2
autocmd FileType elixir setlocal shiftwidth=2 tabstop=2 softtabstop=2
autocmd FileType go setlocal shiftwidth=8 tabstop=8 softtabstop=8 noexpandtab
autocmd BufNewFile,BufRead *.js.es6 set syntax=javascript
set nojoinspaces
set ai  " Auto indent
set si  " Smart indent
set wrap  " Soft-wrap lines
set textwidth=0 wrapmargin=0 " Don't automatically hard-wrap lines.
set encoding=utf8
set ffs=unix,dos,mac
set ignorecase
set smartcase
set hlsearch
set incsearch
set foldmethod=manual
set foldcolumn=1
let g:mapleader = "\<Space>"
let maplocalleader = "\\"
