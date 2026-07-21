import shutil

# UI Palette Constants
RESET = "\033[0m"
BOLD = "\033[1m"
WHITE = "\033[37m"
DIM = "\033[2m"
UNDERLINE = "\033[4m"

BG_CYAN = "\033[46m"
BG_GREEN = "\033[42m"

# Foreground Colors
CYAN = "\033[36m"
GREEN = "\033[32m"
YELLOW = "\033[33m"
MAGENTA = "\033[35m"

# Visual Elements
BANNER_STR = f"""
{CYAN}{BOLD}===================================================={RESET}
{MAGENTA}{BOLD}          ⚡ FLASHCARD CLI - PYTHON ENGINE ⚡          {RESET}
{CYAN}{BOLD}===================================================={RESET}
 {GREEN}➔ Layout Engine Loaded Successfully{RESET}
 {GREEN}➔ Listening for Host Commands...{RESET}
{CYAN}{BOLD}----------------------------------------------------{RESET}
"""

FRONT_INSTRUCTION = f"{YELLOW}{BOLD}[?] Guess{RESET} | {UNDERLINE}Press Enter to Flip the Card {RESET}"
BACK_INSTRUCTION  = f"{GREEN}{BOLD}[!] Nailed it{RESET} | {UNDERLINE}Press key to continue... {RESET}"

class CustomFlashcardRenderer:
    """
    Third-Party Flashcard Renderer Layout.
    Developers modify this class to customize terminal presentations.
    """
    def render_card(self, card, card_num, total_cards, unparsed_modifiers):
        """
        Processes a single flashcard and outputs custom layout formats.
        Returns a tuple of exactly three elements: (front_view, back_view, progress_bar)
        """
        # Determine active terminal window columns to scale full-width banners
        columns, _ = shutil.get_terminal_size(fallback=(80, 24))

        # 1. Format the Question State (Step A)
        front_view = self._build_banner(" FRONT SIDE (QUESTION) ", BG_CYAN, columns)
        front_view += f"\n{BOLD}{card.front}{RESET}\n\n"

        # 2. Format the Answer State (Step B)
        back_view = self._build_banner(" FRONT SIDE (QUESTION) ", BG_CYAN, columns)
        back_view += f"\n{card.front}\n\n"
        back_view += self._build_banner(" BACK SIDE (ANSWER) ", BG_GREEN, columns)
        back_view += f"\n{BOLD}{card.back}{RESET}\n\n"

        # 3. Format the Timeline Indicator (feel free to skip this step to just use default [1/4] type progress indicator)
        progress_bar = f"{DIM}[Progress Metrics: {card_num}/{total_cards}]{RESET}"

        return front_view, back_view, progress_bar

    def _build_banner(self, title_text, bg_color, total_width):
        if len(title_text) > total_width:
            title_text = title_text[:total_width-3] + "..."
            
        remaining_space = total_width - len(title_text)
        left_pad = remaining_space // 2
        right_pad = remaining_space - left_pad

        return f"{bg_color}{WHITE}{BOLD}{' ' * left_pad}{title_text}{' ' * right_pad}{RESET}\n"
