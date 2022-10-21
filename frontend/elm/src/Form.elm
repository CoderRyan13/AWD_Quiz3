module Form exposing (main)

import Html exposing (..)
import Html.Attributes exposing (..)

type alias User =
    { task : String
    }

initialModel : User
initialModel =
    { task = ""
    }

view : User -> Html msg
view user =
    div []
        [ h1 [] [ text "Todo List" ]
        , Html.form []
            [ div []
                [ text "New task"
                , input [ id "task", type_ "text" ] []
                ]
            , div []
                [ button [ type_ "submit" ]
                    [ text "Add Task" ]
                ]
            ]
        ]
main : Html msg
main =
    view initialModel