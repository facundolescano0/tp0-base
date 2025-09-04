class ServerProtocol:
    MAX_LENGTH = 1024
    SIZE = 2
    def __init__(self, client_sock):
        self.client_sock = client_sock
        self.max_length = self.MAX_LENGTH

    def recv_all_bytes(self, bytes_amount):
        buf = bytearray()
        while len(buf) < bytes_amount:
            chunk = self.client_sock.recv(bytes_amount - len(buf))
            if not chunk:
                raise ConnectionError("socket closed while reading")
            buf.extend(chunk)
        return bytes(buf)

    def recv_bet_size(self):
        data = self.recv_all_bytes(ServerProtocol.SIZE)
        num = int.from_bytes(data, 'big')
        return num

    def recv_bet(self):
        size = self.recv_bet_size()
        msg = self.recv_all_bytes(size).decode('utf-8').rstrip()

        campos = msg.split('|')
        if len(campos) == 6:
            return campos
        else:
            return None

    def send_response_bet(self, nid, number):
        response = f"{nid}|{number}\n"
        response_size = len(response)
        data = response_size.to_bytes(self.SIZE, 'big')

        self.client_sock.sendall(data)
        self.client_sock.sendall(response.encode('utf-8'))


    def close(self):
        self.client_sock.close()
