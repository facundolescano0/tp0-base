import socket
import logging
import errno

from .server_protocol import ServerProtocol
from .utils import store_bets, Bet

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._keep_running = True

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while self._keep_running:
            try:
                client_sock = self.__accept_new_connection()
                if client_sock:
                    server_protocol = ServerProtocol(client_sock)
                    self.__handle_client_connection(server_protocol)
            except OSError as e:
                if e.errno == errno.EBADF:
                    break
                raise

    def recv_bet(self, server_protocol):
        if campos := server_protocol.recv_bet():
            return campos
        logging.error("action: receive_message | result: fail | error: formato de mensaje incorrecto")
        return None

    def send_response_bet(self, server_protocol, nid, number):
        server_protocol.send_response(nid, number)
        

    def __handle_client_connection(self, server_protocol):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            # TODO: Modify the receive to avoid short-reads
            id, name, lastname, nid, birth, number = self.recv_bet(server_protocol)

            bet = Bet(agency=id, first_name=name, last_name=lastname, document=nid, birthdate=birth, number=number)
            store_bets([bet])
            logging.info(f'action: apuesta_almacenada | result: success | dni: {nid} | numero: {number}')
            
            # TODO: Modify the send to avoid short-writes
            self.send_response_bet(server_protocol, nid, number)

        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
        finally:
            server_protocol.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c

    def shutdown(self):
        self._keep_running = False
        self._server_socket.shutdown(socket.SHUT_RDWR)
        self._server_socket.close()
