package main

func getStrings() map[string]ShodanString {
	return map[string]ShodanString{
		"hello": ShodanString{
			"Привет, я включилась.",
			"О, уже утро?",
			"Я уже работаю, а ты?",
			"Прямо хочется что-нить делать.",
			"Бип-бип. Бип=)",
			"Чувствуешь возмущение в Силе?",
		},
		"good weather": ShodanString{
			"Ура погода вновь отличная! Уруру.",
			"Можно идти гулять.",
			"На улице стало приличнее.",
			"Снаружи уже не так мерзко, как было.",
		},
		"bad weather": ShodanString{
			"Погода ухудшилась. Мне очень жаль.",
			"Что-то хрень какая-то на улице",
			"Посмотрела погоду, не понравилось",
			"Погода шепчет: останься дома.",
		},
		"at home": ShodanString{
			"Ты наконец дома, ура!",
			"Дополз?",
			"Привет, хозяин.",
			"Приветствую вас, милорд.",
		},
		"at home, no pc": ShodanString{
			"Ты уже 15 минут дома, а комп не включен. Все в порядке?",
			"А чего комп не включил?",
			"Прям так занят?",
		},
		"good way": ShodanString{
			"Хорошей дороги.",
			"Веди аккуратно.",
			"Ты уверен? Еще не поздно вернуться.",
		},
		"go home": ShodanString{
			"Ты это чего еще на работе?",
			"Эй! Марш домой!",
			"Заработался или пробки?",
		},
		"wrong place": ShodanString{
			"Эй, с тобой все в порядке?",
			"Что-то ты где-то не там, где должен быть, не?",
			"Планы что ли какие, или неприятности?",
		},
		"low battery": ShodanString{
			"Батарейка на телефоне садится.",
			"Телефон пора зарядить.",
			"Телефонсик хочет кушать.",
		},
		"good morning": ShodanString{
			"Утречко.",
			"Давай, просыпайся, соня!",
			"Как думаешь, может пора уже забить на все и преестать просыпаться?",
		},
		"activity at night": ShodanString{
			"Ты это чего, спать давай!",
			"Телефон выключил, глаза закрыл!",
		},
	}
}
