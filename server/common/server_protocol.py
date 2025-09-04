import socket
import logging


class ServerProtocol:
    BATCH_FINISHED = 0
    ONE_BYTE = 1
    SIZE = 2
    BET_FIELDS = 6

    # server to client
    BATCH_OK = 3
    BATCH_FAIL = 4

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
    
    def recv_batch_size(self):
        data = self.recv_all_bytes(ServerProtocol.SIZE)
        num = int.from_bytes(data, 'big')
        return num
    
    def recv_batch(self):
        bytes_amount = self.recv_batch_size()
        if not bytes_amount:
            return None

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
            campos = line.split('|')
            if len(campos) == ServerProtocol.BET_FIELDS:
                batch.append(campos)
            else:
                return None
        return batch

    def send_response_batch(self, result):
        self.client_sock.sendall(bytes([result]))
        

    def close(self):
        self.client_sock.close()
