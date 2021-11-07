# Sistemas Distribuídos - Tarea 2

## Integrantes:
* Cristóbal Abarca, 201573060-1
* Maximiliano Bombin, 201104308-1
* Ignacio Rojas, 201403027-4

---------------------------------------------------
# Instrucciones

## Iniciar Procesos:
1. En cada máquina iniciando sesión como el usuario 'alumno', deben ingresar a la carpeta llamada 'Lab2INF343' de la siguiente forma:

    cd Lab2INF343


2. Luego, en esta carpeta se pueden ejecutar los comandos correspondientes a cada máquina. Iniciar los procesos en su máquina correspondiente:
    * Máquina dis149 - Lider
        
            make lider
        
    * Máquina dist150 - NameNode
        
            make namenode
         
    * Máquina dist151 - Pozo
        
            make pozo
         
    * Desde cualquier máquina - Jugador Real 
        
            make jugadorH
         
    * Desde cualquier máquina - Jugadores Bot 
        
            make jugadorB
         
    * Máquina dis149 - DataNode1
        
            make datanode1
        
    * Máquina dis151 - DataNode2
        
            make datanode2
        
    * Máquina dis152 - DataNode3
        
            make datanode3
        

3. Con make clean se limpian los .txt en caso de querer realizar otra ejecución.
---------------------------------------------------
## Jugar:
1. Inicializar NameNode.
2. Inicializar los DataNodes (en las 3 máquinas).
3. Inicializar Pozo.
4. Inicializar Líder.
5. Inicializar Jugadores (Los 16, pueden ser humanos o bots(make jugadorH o make jugadorB respectivamente)).
6. Cada jugador humano(real, NO bot) debe apretar "s" o "S" + "ENTER" antes de que el líder inicie el juego.
7. Cuando los 16 jugadores se unan, el líder debe apretar "s" o "S" + "ENTER" para dar comienzo al juego.
8. A continuación Los jugadores reales deben apretar "s" o "S" + "ENTER" y se les preguntará un número para jugar dependiendo el juego en el que está.
9. El líder pasará a una siguiente ronda ("s" + "ENTER") o juego/etapa solo cuando ya se enviaron todas las respuestas de los clientes(los bots lo hacen automático, pero los reales tienen que escribir el número por consola).
10. El jugador humano (jugadorH) puede consultar el monto acumulado presionado "m" o "M" + "ENTER"
11. Cuando se tengan jugadas, el Líder puede pedir las asociadas a un jugador, presionando "j" o "J" + "ENTER" y luego especificar el número del jugador deseado.