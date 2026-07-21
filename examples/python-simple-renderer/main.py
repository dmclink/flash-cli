#!/usr/bin/env python3
import sys
import os
from concurrent import futures
import grpc

# Core plumbing: Automatically wires up the bundled gRPC protobuf stubs.
CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
# modify this path if you bundle the proto stubs in a different location than in this example
PY_GEN_DIR = os.path.join(CURRENT_DIR, "gen", "python")

if os.path.exists(PY_GEN_DIR):
    if PY_GEN_DIR not in sys.path:
        sys.path.insert(0, PY_GEN_DIR)
else:
    print(f"Error: Could not locate bundled proto stubs folder at: {PY_GEN_DIR}", file=sys.stderr)
    sys.exit(1)

from render.v1 import render_pb2
from render.v1 import render_pb2_grpc

# This pulls in your custom styling layout code. 
# to build your own renderer plugin, leave this main.py alone and edit renderer.py.
# edit the import class as necessary
from renderer import CustomFlashcardRenderer
from renderer import BANNER_STR
from renderer import BACK_INSTRUCTION
from renderer import FRONT_INSTRUCTION


# Network adapter: Translates incoming gRPC network requests from the core app
# and forwards the arguments directly into your custom renderer.py script.
class RenderServiceRouter(render_pb2_grpc.RenderServiceServicer):
    def __init__(self, user_renderer, grpc_server):
        self.user_renderer = user_renderer
        self.grpc_server = grpc_server

    def Process(self, request, context):
        try:
            # Map request variables directly into your render_card arguments.
            # If you want to change what your plugin accepts, you must update the arguments
            # inside CustomFlashcardRenderer.render_card() to match this line exactly.
            # double check render.proto to ensure request variable fields match the proto object
            # feel free not to use unparsed_modifiers if your plugin doesn't support any custom mods
            front, back, progress = self.user_renderer.render_card(
                card=request.card,
                card_num=request.current_card_num,
                total_cards=request.total_card_count,
                unparsed_modifiers=list(request.unparsed_modifiers) if hasattr(request, 'unparsed_modifiers') else []
            )
            
            # fields here must match. front and back must be included for obvious reasons. 
            # progress has a sane default on empty strings and can safely be dropped. 
            # If no progress bar desired just pass a string with one whitespace like " "
            return render_pb2.ProcessResponse(
                formatted_front=front,
                formatted_back=back,
                progress=progress
            )
        except Exception as e:
            # Captures any crashes inside your renderer.py and pipes the stack trace
            # safely to the main console stream so you can debug your layout errors.
            print(f"Plugin error executing render custom logic: {e}", file=sys.stderr)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            raise e

    def Init(self, request, context):
        try:
            # Satisfies the Init implementation signature mapping back metadata to Go host
            return render_pb2.InitResponse(
                startup_banner=BANNER_STR,
                instruction_front=FRONT_INSTRUCTION,
                instruction_back=BACK_INSTRUCTION
            )
        except Exception as e:
            print(f"Plugin error executing initialization sequence: {e}", file=sys.stderr)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            raise e

    def Shutdown(self, request, context):
        print("Received graceful shutdown command from host.", file=sys.stderr)
        self.grpc_server.stop(grace=0.5)
        return render_pb2.ShutdownResponse()


# Initialization engine: Boots up the background gRPC network bus daemon
# and negotiates the required subprocess handshake parameters with the Go host.
def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
    
    user_implementation = CustomFlashcardRenderer()
    router = RenderServiceRouter(user_implementation, server)
    
    render_pb2_grpc.add_RenderServiceServicer_to_server(router, server)

    port = server.add_insecure_port('127.0.0.1:0')
    server.start()

    if os.environ.get("FLASHCARD_CLI_PLUGIN_HANDSHAKE") != "flashcards-grpc-ecosystem-auth":
        print("Insecure authentication handshake failure.", file=sys.stderr)
        sys.exit(1)

    # Core protocol link: Never print() or write anywhere else to sys.stdout in your code.
    # The main application hooks into this exact standard output stream line to connect.
    sys.stdout.write(f"1|1|tcp|127.0.0.1:{port}|grpc\n")
    sys.stdout.flush()

    server.wait_for_termination()


if __name__ == '__main__':
    serve()
