return {
  {
    "GCBallesteros/jupytext.nvim",
    lazy = false, -- load at startup so it can catch .ipynb files when they open
    opts = {
      style = "markdown",
      output_extension = "md",
      force_ft = "markdown",
    },
  },
}
