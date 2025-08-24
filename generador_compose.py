class GeneradorCompose:
    def generar_nombre(self, f):
        f.write("name: tp0\n")

    def generar_server(self, f):

        f.write("  server:\n")
        f.write("    container_name: server\n")
        f.write("    image: server:latest\n")
        f.write("    entrypoint: python3 /main.py\n")
        f.write("    environment:\n")
        f.write("      - PYTHONUNBUFFERED=1\n")
        f.write("      - LOGGING_LEVEL=DEBUG\n")
        f.write("    networks:\n")
        f.write("      - testing_net\n")

    def generar_clientes(self, f, cantidad_clientes):
        for i in range(cantidad_clientes):
            f.write(f"  cliente_{i}:\n")
            f.write(f"    container_name: cliente_{i}\n")
            f.write(f"    image: cliente:latest\n")
            f.write(f"    entrypoint: /client\n")
            f.write(f"    environment:\n")
            f.write(f"      - CLI_ID={i}\n")
            f.write(f"      - CLI_LOG_LEVEL=DEBUG\n")
            f.write(f"    networks:\n")
            f.write(f"      - testing_net\n")
            f.write(f"    depends_on:\n")
            f.write(f"      - server\n")

    def generar_servicios(self, f, cantidad_clientes):

        f.write("services:\n")
        self.generar_server(f)
        self.generar_clientes(f, cantidad_clientes)


    def generar_redes(self, f):
        f.write(f"  networks:\n")
        f.write(f"    testing_net:\n")
        f.write(f"      ipam:\n")
        f.write(f"        driver: default\n")
        f.write(f"        config:\n")
        f.write(f"          - subnet: 172.25.125.0/24\n")


    def generar_compose(self, nombre_archivo, cantidad_clientes):
        with open(nombre_archivo, 'w') as f:
            self.generar_nombre(f)
            self.generar_servicios(f, cantidad_clientes)
            self.generar_redes(f)
        f.close()

if __name__ == "__main__":
    import sys
    if len(sys.argv) != 3:
        print("Uso: python3 generador_compose.py <nombre_archivo> <cantidad_clientes>")
        sys.exit(1)
    nombre_archivo = sys.argv[1]
    cantidad_clientes = int(sys.argv[2])
    generador = GeneradorCompose()
    generador.generar_compose(nombre_archivo, cantidad_clientes)