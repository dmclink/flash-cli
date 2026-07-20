#!/usr/bin/env python3
import sys
import os
import shutil
from concurrent import futures
import grpc

# find the bundled 'gen/python' directory right next to this script
CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
PY_GEN_DIR = os.path.join(CURRENT_DIR, "gen", "python")

if os.path.exists(PY_GEN_DIR):
    if PY_GEN_DIR not in sys.path:
        sys.path.insert(0, PY_GEN_DIR)
else:
    print(f"Error: Could not locate bundled proto stubs folder at: {PY_GEN_DIR}", file=sys.stderr)
    sys.exit(1)

from render.v1 import render_pb2
from render.v1 import render_pb2_grpc

RESET = "\033[0m"
BOLD = "\033[1m"
WHITE = "\033[37m"
DIM = "\033[2m"
YELLOW = "\033[33m"

BG_CYAN = "\033[46m"
BG_GREEN = "\033[42m"


class SimpleVerticalRenderer(render_pb2_grpc.RenderServiceServicer):
    """
    Implements the RenderService contract. Returns stacked vertical layouts
    that are easily read by the host application's printing systems.
    """
    def Process(self, request, context):
        card = request.card
        card_num = request.current_card_num
        total_cards = request.total_card_count

        columns, _ = shutil.get_terminal_size(fallback=(80, 24))

        front_view = self.build_banner(" FRONT SIDE (QUESTION) ", BG_CYAN, columns)
        front_view += f"\n{BOLD}{card.front}{RESET}\n\n"

        back_view = self.build_banner(" FRONT SIDE (QUESTION) ", BG_CYAN, columns)
        back_view += f"\n{card.front}\n\n"
        back_view += self.build_banner(" BACK SIDE (ANSWER) ", BG_GREEN, columns)
        back_view += f"\n{BOLD}{card.back}{RESET}\n\n"

        progress_bar = f"{DIM}[Progress Metrics: {card_num}/{total_cards}]{RESET}"

        return render_pb2.ProcessResponse(
            formatted_front=front_view,
            formatted_back=back_view,
            progress=progress_bar
        )

    def build_banner(self, title_text, bg_color, total_width):
        """
        Creates a full-width colored header line with centered title text.
        """
        if len(title_text) > total_width:
            title_text = title_text[:total_width-3] + "..."
            
        remaining_space = total_width - len(title_text)
        left_pad = remaining_space // 2
        right_pad = remaining_space - left_pad

        return f"{bg_color}{WHITE}{BOLD}{' ' * left_pad}{title_text}{' ' * right_pad}{RESET}\n"


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=5))
    render_pb2_grpc.add_RenderServiceServicer_to_server(SimpleVerticalRenderer(), server)
    
    port = server.add_insecure_port('127.0.0.1:0')
    server.start()

    if os.environ.get("FLASHCARD_CLI_PLUGIN_HANDSHAKE") != "flashcards-grpc-ecosystem-auth":
        print("Insecure authentication handshake failure.", file=sys.stderr)
        sys.exit(1)

        # don't forget this line!
    sys.stdout.write(f"1|1|tcp|127.0.0.1:{port}|grpc\n")
    sys.stdout.flush()

    server.wait_for_termination()


if __name__ == '__main__':
    serve()
