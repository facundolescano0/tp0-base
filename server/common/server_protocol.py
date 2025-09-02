import logging
import socket


class ServerProtocol:
    def __init__(self, client_sock):
        self.client_sock = client_sock
    
    def recv_all(self):
        max_length = 8192
        buffer = b""
        while True:
            chunk = self.client_sock.recv(max_length-len(buffer))
            if not chunk:
                break 
            buffer += chunk
            if b"\n" in chunk:
                break
        msg = buffer.rstrip().decode('utf-8')
        return msg

    def recv_bet(self):
        msg = self.recv_all()
        campos = msg.split('|')
        if len(campos) == 6:
            return campos
        else:
            return None

    def send_response(self, nid, number):
        response = f"{nid}|{number}\n"
        self.client_sock.sendall(response.encode('utf-8'))

    def recv_batch(self):
        data = self.recv_all()
        if not data:
            return None
        lines = data.splitlines()
        if not lines:
            return None
        header = lines[0]

        try:
            logging.info(f"intento convertir a int")
            bet_count = int(header.strip())
        except ValueError:
            logging.error("action: receive_message | result: fail | error: no puedo convertir a int")
            return None
        
        if len(lines[1:]) < bet_count:
            logging.error("No se recibieron todas las apuestas esperadas")
            return None

        batch = []
        for line in lines[1:1+bet_count]:
            campos = line.split('|')
            if len(campos) == 6:
                batch.append(campos)
            else:
                return None
            
        return batch

    def send_response_batch(self, message):
        response = f"{message}\n"
        self.client_sock.sendall(response.encode('utf-8'))
        

    def close(self):
        self.client_sock.shutdown(socket.SHUT_RDWR)
        self.client_sock.close()
