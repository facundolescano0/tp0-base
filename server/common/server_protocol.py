import socket
import logging


class ServerProtocol:
   
    BATCH_FINISHED = 0
    ONE_BYTE = 1
    SIZE = 2
   
    # client to server
    BATCH = 1
    WINNERS_REQUEST = 2

    # server to client
    BATCH_OK = 3
    BATCH_FAIL = 4
    SEND_WINNERS = 5
    NOT_READY = 6
    NO_WINNERS = 0


    def __init__(self, client_sock, max_length):
        self.client_sock = client_sock
        self.max_length = max_length

    def recv_all_bytes(self, bytes_amount):
        buf = bytearray()
        while len(buf) < bytes_amount:
            chunk = self.client_sock.recv(bytes_amount - len(buf))
            if not chunk:
                raise ConnectionError("socket closed while reading")
            buf.extend(chunk)
        return bytes(buf)

    def recv_opcode(self):
        buffer = self.recv_all_bytes(ServerProtocol.ONE_BYTE)
        if buffer:
            return buffer[0]
        return None
    
    def recv_batch_size(self):
        data = self.recv_all_bytes(ServerProtocol.SIZE)
        num = int.from_bytes(data, 'big')
        return num

    def recv_batch(self):
        bytes_amount = self.recv_batch_size()
        if bytes_amount == ServerProtocol.BATCH_FINISHED:
            return ServerProtocol.BATCH_FINISHED

        if bytes_amount < 0 or bytes_amount > self.max_length:
            return None

        data = self.recv_all_bytes(bytes_amount)
        try:
            text = data.decode('utf-8')
        except UnicodeDecodeError:
            logging.info("invalid UTF-8 in batch payload")
            return None

        lines = text.splitlines()
        if not lines:
            return None

        batch = []
        for line in lines:
            fields = line.split('|')
            if len(fields) == 6:
                batch.append(fields)
            else:
                return None
        return batch

    def recv_agency_id(self):
        buffer = self.recv_all_bytes(ServerProtocol.ONE_BYTE)
        if buffer:
            return buffer[0]
        return None

    def send_response_batch(self, result):
        self.client_sock.sendall(bytes([result]))


    def send_winners(self, winners):

        if len(winners) == 0:
            self.client_sock.sendall(bytes([self.SEND_WINNERS]))
            self.client_sock.sendall((self.NO_WINNERS).to_bytes(ServerProtocol.SIZE, 'big'))
            return

        msg = ""
        first = True
        for winner in winners:
            if not first:
                msg += "|"
            msg += winner
            first = False

        payload = msg.encode('utf-8')

        self.client_sock.sendall(bytes([self.SEND_WINNERS]))

        size_winners = len(payload)
        self.client_sock.sendall(size_winners.to_bytes(ServerProtocol.SIZE, 'big'))

        self.client_sock.sendall(payload)

    def send_not_ready(self):
        self.client_sock.sendall(bytes([ServerProtocol.NOT_READY]))

    def close(self):
        self.client_sock.close()

