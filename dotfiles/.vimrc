" Some settings to make Vim nicer on remote servers
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
set softtabstop=2
set shiftwidth=2
set tabstop=2
set expandtab
autocmd FileType python setlocal shiftwidth=4 tabstop=4 softtabstop=4
autocmd FileType rust setlocal shiftwidth=4 tabstop=4 softtabstop=4
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
