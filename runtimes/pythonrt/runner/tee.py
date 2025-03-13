from queue import Queue
from threading import Thread
from time import sleep
import grpc
import pb


class Tee:
    def __init__(self, stream, runner_id, worker):
        self.stream = stream
        self.queue = Queue()
        self.running = True
        self.worker = worker
        self.runner_id = runner_id

        Thread(target=self.print_thread, daemon=True).start()

    def close(self):
        self.running = False
        self.stream.close()

    def write(self, data):
        self.stream.write(data)
        self.queue.put(data)

    def flush(self):
        self.stream.flush()
        while not self.queue.empty():
            sleep(0.001)

    def print_thread(self):
        while self.running:
            data = self.queue.get()
            self.rpc_call(data)

    def rpc_call(self, data):
        if not data.strip():
            return

        req = pb.handler.PrintRequest(
            runner_id=self.runner_id,
            message=data,
        )

        try:
            self.worker.Print(req)
        except grpc.RpcError as err:
            if err.code() in (grpc.StatusCode.UNAVAILABLE, grpc.StatusCode.CANCELLED):
                print("gRPC connection unavailable, print failed", file=self.stream)
            else:
                print(f"Print RPC error: {err}", file=self.stream)
