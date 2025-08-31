class ServerProtocol:
    def __init__(self, client_sock):
        self.client_sock = client_sock

    def recv_bet(self):
        buffer = b""
        while True:
            chunk = self.client_sock.recv(1024)
            if not chunk:
                break 
            buffer += chunk
            if b"\n" in chunk:
                break
        msg = buffer.rstrip().decode('utf-8')
        campos = msg.split('|')
        if len(campos) == 6:
            return campos
        else:
            return None

    def send_response(self, nid, number):
        response = f"{nid}|{number}\n"
        self.client_sock.send(response.encode('utf-8'))

    def close(self):
        self.client_sock.close()
